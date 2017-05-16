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

	"github.com/github/hub/github"
	"github.com/mitchellh/go-homedir"
	"github.com/octokit/go-octokit/octokit"
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
		log.Printf("can't detect remote repository: %#v\n", err)
		return exitError
	}
	proj, err := remotes[0].Project()
	if err != nil {
		log.Printf("failed to retrieve project: %#v\n", err)
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
		log.Printf("failed to `gobump show`: %#v\n", err)
		return exitError
	}

	var v struct {
		Version string `json:"version"`
	}
	err = json.Unmarshal(out, &v)
	if err != nil {
		log.Printf("failed to unmarshal `gobump show`'s output: %#v\n", err)
		return exitError
	}
	err = uploadToGithubRelease(proj, v.Version, *staging, *dryRun)
	if err != nil {
		log.Printf("error occured while uploading artifacts to github: %#v\n", err)
		return exitError
	}
	return exitOK
}

var errAlreadyReleased = fmt.Errorf("the release of this version has already existed at GitHub Relase, so skip the process")

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
			log.Println(err.Error())
			return nil
		}
		return err
	}

	body := pr.Body
	assets, err := collectAssets()
	if err != nil {
		return fmt.Errorf("error occured while collecting releasing assets: %#v", err)
	}

	host, err := github.CurrentConfig().PromptForHost(proj.Host)
	if err != nil {
		return fmt.Errorf("failed to detect github config: %#v", err)
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
			return fmt.Errorf("failed to create release: %#v", err)
		}

		err = uploadAssets(gh, release, assets)
		if err != nil {
			return err
		}
		if !staging {
			release, err = gh.EditRelease(release, map[string]interface{}{
				"prerelease": false,
			})
		}
	}
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
		return nil, fmt.Errorf("failed to detect release pull request: %#v", r.Err)
	}
	return &prs[0], nil
}

func handleOldRelease(octoCli *octokit.Client, owner, repo, tag string, staging, dryRun bool) error {
	releaseByTagURL := octokit.Hyperlink("repos/{owner}/{repo}/releases/tags/{tag}")
	u, err := releaseByTagURL.Expand(octokit.M{"owner": owner, "repo": repo, "tag": tag})
	if err != nil {
		return fmt.Errorf("failed to build GitHub URL: %#v", err)
	}
	release, r := octoCli.Releases(u).Latest()
	if r.Err != nil {
		rerr, ok := r.Err.(*octokit.ResponseError)
		if !ok {
			return fmt.Errorf("failed to fetch release: %#v", r.Err)
		}
		if rerr.Response == nil || rerr.Response.StatusCode != http.StatusNotFound {
			return fmt.Errorf("failed to fetch release: %#v", r.Err)
		}
	}
	if release != nil {
		if !staging {
			return errAlreadyReleased
		}
		if !dryRun {
			req, err := octoCli.NewRequest(release.URL)
			if err != nil {
				return fmt.Errorf("something went wrong: %#v", err)
			}
			sawyerResp := req.Request.Delete()
			if sawyerResp.IsError() {
				return fmt.Errorf("release deletion unsuccesful, %#v", sawyerResp.ResponseError)
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
		_, err := gh.UploadReleaseAsset(release, asset, "")
		if err != nil {
			return fmt.Errorf("failed to upload asset: %s, error: %#v", asset, err)
		}
	}
	return nil
}
