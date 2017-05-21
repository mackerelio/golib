package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/Songmu/retry"
	"github.com/github/hub/github"
	"github.com/mitchellh/go-homedir"
	"github.com/octokit/go-octokit/octokit"
	"github.com/pkg/errors"
)

const (
	exitOK = iota
	exitError
)

const version = "0.0.0"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(argv []string) int {
	remotes, err := github.Remotes()
	if err != nil || len(remotes) < 1 {
		log.Printf("can't detect remote repository: %+v\n", err)
		return exitError
	}
	proj, err := remotes[0].Project()
	if err != nil {
		log.Printf("failed to retrieve project: %+v\n", err)
		return exitError
	}
	fs := flag.NewFlagSet("mackerel-github-release", flag.ContinueOnError)
	var (
		dryRun  = fs.Bool("dry-run", false, "dry-run mode")
		staging = fs.Bool("staging", false, "staging release")
	)
	err = fs.Parse(argv)
	if err != nil {
		if err == flag.ErrHelp {
			return exitOK

		}
		return exitError
	}

	out, err := exec.Command("gobump", "show").Output()
	if err != nil {
		log.Printf("failed to `gobump show`: %+v\n", err)
		return exitError
	}

	var v struct {
		Version string `json:"version"`
	}
	err = json.Unmarshal(out, &v)
	if err != nil {
		log.Printf("failed to unmarshal `gobump show`'s output: %+v\n", err)
		return exitError
	}
	log.Printf("Start uploading files to GitHub Releases. version: %s, staging: %t, dry-run: %t\n", v.Version, *staging, *dryRun)
	err = uploadToGithubRelease(proj, v.Version, *staging, *dryRun)
	if err != nil {
		log.Printf("error occured while uploading artifacts to github: %+v\n", err)
		return exitError
	}
	return exitOK
}

var errAlreadyReleased = fmt.Errorf("the release of this version has already existed at GitHub Releases, so skip the process")

func uploadToGithubRelease(proj *github.Project, releaseVer string, staging, dryRun bool) error {
	tag := "staging"
	if !staging {
		tag = "v" + releaseVer
	}
	repo, owner := proj.Name, proj.Owner
	octoCli := getOctoCli()

	pr, err := getReleasePullRequest(octoCli, owner, repo, releaseVer)
	if err != nil {
		return err
	}

	err = handleOldRelease(octoCli, owner, repo, tag, staging, dryRun)
	if err != nil {
		if err == errAlreadyReleased {
			log.Printf("%s. version: %s\n", err, tag)
			return nil
		}
		return err
	}

	body := pr.Body
	assets, err := collectAssets()
	if err != nil {
		return errors.Wrap(err, "error occured while collecting releasing assets")
	}
	sort.Strings(assets)
	log.Println("uploading following files:")
	for _, f := range assets {
		log.Printf("- %s\n", f)
	}

	host, err := github.CurrentConfig().PromptForHost(proj.Host)
	if err != nil {
		return errors.Wrap(err, "failed to detect github config")
	}
	gh := github.NewClientWithHost(host)

	if !dryRun {
		params := &github.Release{
			TagName:    tag,
			Name:       tag,
			Body:       body,
			Prerelease: true,
		}
		release, err := gh.CreateRelease(proj, params)
		if err != nil {
			return errors.Wrap(err, "failed to create release")
		}

		err = uploadAssets(gh, release, assets)
		if err != nil {
			return err
		}
		if !staging {
			err = retry.Retry(3, 3*time.Second, func() error {
				_, err := gh.EditRelease(release, map[string]interface{}{
					"prerelease": false,
				})
				if err != nil {
					log.Println(err)
				}
				return err
			})
			if err != nil {
				return errors.Wrapf(
					err,
					"Upload done, but failed to update prerelease status from true to false. You can check the status and update manually. version: %s",
					releaseVer)
			}
		}
	}
	log.Printf("Upload done. version: %s, staging: %t, dry-run: %t\n", releaseVer, staging, dryRun)
	return nil
}

func getOctoCli() *octokit.Client {
	var auth octokit.AuthMethod
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		auth = octokit.TokenAuth{AccessToken: token}
	}
	return octokit.NewClient(auth)
}

func getReleasePullRequest(octoCli *octokit.Client, owner, repo, releaseVer string) (*octokit.PullRequest, error) {
	releaseBranch := "bump-version-" + releaseVer
	u, err := octokit.PullRequestsURL.Expand(octokit.M{"owner": owner, "repo": repo})
	if err != nil {
		return nil, fmt.Errorf("something went wrong while expanding pullrequest url")
	}
	q := u.Query()
	q.Set("state", "closed")
	q.Set("head", fmt.Sprintf("%s:%s", owner, releaseBranch))
	u.RawQuery = q.Encode()
	prs, r := octoCli.PullRequests(u).All()
	if r.HasError() || len(prs) != 1 {
		return nil, fmt.Errorf("failed to detect release pull request: %+v", r.Err)
	}
	return &prs[0], nil
}

func handleOldRelease(octoCli *octokit.Client, owner, repo, tag string, staging, dryRun bool) error {
	releaseByTagURL := octokit.Hyperlink("repos/{owner}/{repo}/releases/tags/{tag}")
	u, err := releaseByTagURL.Expand(octokit.M{"owner": owner, "repo": repo, "tag": tag})
	if err != nil {
		return errors.Wrap(err, "failed to build GitHub URL")
	}
	release, r := octoCli.Releases(u).Latest()
	if r.Err != nil {
		rerr, ok := r.Err.(*octokit.ResponseError)
		if !ok || rerr.Response == nil || rerr.Response.StatusCode != http.StatusNotFound {
			return errors.Wrap(r.Err, "failed to fetch release")
		}
	}
	if release != nil {
		if !staging {
			return errAlreadyReleased
		}
		if !dryRun {
			req, err := octoCli.NewRequest(release.URL)
			if err != nil {
				return errors.Wrap(err, "something went wrong")
			}
			sawyerResp := req.Request.Delete()
			if sawyerResp.IsError() {
				return errors.Wrap(sawyerResp.ResponseError, "release detection unsuccesful")
			}
			defer sawyerResp.Body.Close()

			if sawyerResp.StatusCode != http.StatusNoContent {
				return fmt.Errorf("could not delete the release corresponding to tag %s", tag)
			}
		}
	}
	return nil
}

func collectAssets() (assets []string, err error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	for _, glob := range [...]string{
		home + "/rpmbuild/RPMS/*/*.rpm",
		"rpmbuild/RPMS/*/*.rpm",
		"packaging/*.deb",
		"snapshot/*.zip",
		"snapshot/*.tar.gz",
		"build/*.tar.gz",
	} {
		files, err := filepath.Glob(glob)
		if err != nil {
			return nil, err
		}
		assets = append(assets, files...)
	}
	return assets, nil
}

func uploadAssets(gh *github.Client, release *github.Release, assets []string) error {
	for _, asset := range assets {
		err := retry.Retry(3, 3*time.Second, func() error {
			_, err := gh.UploadReleaseAsset(release, asset, "")
			if err != nil {
				log.Printf("failed to upload asset: %s, error: %+v", asset, err)
			}
			return err
		})
		if err != nil {
			return errors.Wrapf(err, "failed to upload asset and gave up: %s", asset)
		}
	}
	return nil
}
