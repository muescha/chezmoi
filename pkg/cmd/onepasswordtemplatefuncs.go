package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoilog"
)

var onepasswordVersionRx = regexp.MustCompile(`^(\d+\.\d+\.\d+\S*)`)

type onepasswordConfig struct {
	Command     string
	version     *semver.Version
	versionErr  error
	outputCache map[string][]byte
}

type onepasswordArgs struct {
	item    string
	vault   string
	account string
	args    []string
}

type onepasswordItemV1 struct {
	Details struct {
		Fields   []map[string]any `json:"fields"`
		Sections []struct {
			Fields []map[string]any `json:"fields,omitempty"`
		} `json:"sections"`
	} `json:"details"`
}

type onepasswordItemV2 struct {
	Fields []map[string]any `json:"fields"`
}

func (c *Config) onepasswordTemplateFunc(userArgs ...string) map[string]any {
	version, err := c.onepasswordVersion()
	if err != nil {
		panic(err)
	}

	var baseArgs []string
	switch {
	case version.Major == 1:
		baseArgs = []string{"get", "item"}
	case version.Major >= 2:
		baseArgs = []string{"item", "get", "--format", "json"}
	default:
		panic(&unsupportedVersionError{
			version: version,
		})
	}

	args, err := newOnepasswordArgs(baseArgs, userArgs)
	if err != nil {
		panic(err)
	}

	output, err := c.onepasswordOutput(args)
	if err != nil {
		panic(err)
	}

	var data map[string]any
	if err := json.Unmarshal(output, &data); err != nil {
		panic(newParseCmdOutputError(c.Onepassword.Command, args.args, output, err))
	}
	return data
}

func (c *Config) onepasswordDetailsFieldsTemplateFunc(userArgs ...string) map[string]any {
	version, err := c.onepasswordVersion()
	if err != nil {
		panic(err)
	}

	switch {
	case version.Major == 1:
		item, err := c.onepasswordItemV1(userArgs)
		if err != nil {
			panic(err)
		}

		result := make(map[string]any)
		for _, field := range item.Details.Fields {
			if designation, ok := field["designation"].(string); ok {
				result[designation] = field
			}
		}
		return result

	case version.Major >= 2:
		item, err := c.onepasswordItemV2(userArgs)
		if err != nil {
			panic(err)
		}

		result := make(map[string]any)
		for _, field := range item.Fields {
			if _, ok := field["section"]; ok {
				continue
			}
			if id, ok := field["id"].(string); ok && id != "" {
				result[id] = field
				continue
			}
			if label, ok := field["label"].(string); ok && label != "" {
				result[label] = field
				continue
			}
		}
		return result

	default:
		panic(&unsupportedVersionError{
			version: version,
		})
	}
}

func (c *Config) onepasswordDocumentTemplateFunc(userArgs ...string) string {
	version, err := c.onepasswordVersion()
	if err != nil {
		panic(err)
	}

	var baseArgs []string
	switch {
	case version.Major == 1:
		baseArgs = []string{"get", "document"}
	case version.Major >= 2:
		baseArgs = []string{"document", "get"}
	default:
		panic(&unsupportedVersionError{
			version: version,
		})
	}

	args, err := newOnepasswordArgs(baseArgs, userArgs)
	if err != nil {
		panic(err)
	}

	output, err := c.onepasswordOutput(args)
	if err != nil {
		panic(err)
	}
	return string(output)
}

func (c *Config) onepasswordItemFieldsTemplateFunc(userArgs ...string) map[string]any {
	version, err := c.onepasswordVersion()
	if err != nil {
		panic(err)
	}

	switch {
	case version.Major == 1:
		item, err := c.onepasswordItemV1(userArgs)
		if err != nil {
			panic(err)
		}

		result := make(map[string]any)
		for _, section := range item.Details.Sections {
			for _, field := range section.Fields {
				if t, ok := field["t"].(string); ok {
					result[t] = field
				}
			}
		}
		return result

	case version.Major >= 2:
		item, err := c.onepasswordItemV2(userArgs)
		if err != nil {
			panic(err)
		}

		result := make(map[string]any)
		for _, field := range item.Fields {
			if _, ok := field["section"]; !ok {
				continue
			}
			if label, ok := field["label"].(string); ok {
				result[label] = field
			}
		}
		return result

	default:
		panic(&unsupportedVersionError{
			version: version,
		})
	}
}

func (c *Config) onepasswordItemV1(userArgs []string) (*onepasswordItemV1, error) {
	args, err := newOnepasswordArgs([]string{"get", "item"}, userArgs)
	if err != nil {
		return nil, err
	}

	output, err := c.onepasswordOutput(args)
	if err != nil {
		return nil, err
	}

	var item onepasswordItemV1
	if err := json.Unmarshal(output, &item); err != nil {
		return nil, newParseCmdOutputError(c.Onepassword.Command, args.args, output, err)
	}
	return &item, nil
}

func (c *Config) onepasswordItemV2(userArgs []string) (*onepasswordItemV2, error) {
	args, err := newOnepasswordArgs([]string{"item", "get", "--format", "json"}, userArgs)
	if err != nil {
		return nil, err
	}

	output, err := c.onepasswordOutput(args)
	if err != nil {
		return nil, err
	}

	var item onepasswordItemV2
	if err := json.Unmarshal(output, &item); err != nil {
		return nil, newParseCmdOutputError(c.Onepassword.Command, args.args, output, err)
	}
	return &item, nil
}

func (c *Config) onepasswordOutput(args *onepasswordArgs) ([]byte, error) {
	key := strings.Join(args.args, "\x00")
	if output, ok := c.Onepassword.outputCache[key]; ok {
		return output, nil
	}

	//nolint:gosec
	cmd := exec.Command(c.Onepassword.Command, args.args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.Onepassword.outputCache == nil {
		c.Onepassword.outputCache = make(map[string][]byte)
	}
	c.Onepassword.outputCache[key] = output

	return output, nil
}

func (c *Config) onepasswordReadTemplateFunc(url string, args ...string) string {
	onepasswordArgs := &onepasswordArgs{
		args: []string{"read", url},
	}
	switch len(args) {
	case 0:
		// Do nothing.
	case 1:
		onepasswordArgs.args = append(onepasswordArgs.args, "--account", args[0])
	default:
		panic(fmt.Errorf("expected 1 or 2 arguments, got %d", len(args)))
	}

	output, err := c.onepasswordOutput(onepasswordArgs)
	if err != nil {
		panic(err)
	}
	return string(output)
}

func (c *Config) onepasswordVersion() (*semver.Version, error) {
	if c.Onepassword.version != nil || c.Onepassword.versionErr != nil {
		return c.Onepassword.version, c.Onepassword.versionErr
	}

	args := &onepasswordArgs{
		args: []string{"--version"},
	}
	output, err := c.onepasswordOutput(args)
	if err != nil {
		c.Onepassword.versionErr = err
		return nil, c.Onepassword.versionErr
	}

	m := onepasswordVersionRx.FindSubmatch(output)
	if m == nil {
		c.Onepassword.versionErr = &extractVersionError{
			output: output,
		}
		return nil, c.Onepassword.versionErr
	}

	version, err := semver.NewVersion(string(m[1]))
	if err != nil {
		c.Onepassword.versionErr = &parseVersionError{
			output: m[1],
			err:    err,
		}
		return nil, c.Onepassword.versionErr
	}

	c.Onepassword.version = version
	return c.Onepassword.version, c.Onepassword.versionErr
}

func newOnepasswordArgs(baseArgs, userArgs []string) (*onepasswordArgs, error) {
	if len(userArgs) < 1 || 3 < len(userArgs) {
		return nil, fmt.Errorf("expected 1, 2, or 3 arguments, got %d", len(userArgs))
	}

	a := &onepasswordArgs{
		args: baseArgs,
	}
	a.item = userArgs[0]
	a.args = append(a.args, a.item)
	if len(userArgs) > 1 && userArgs[1] != "" {
		a.vault = userArgs[1]
		a.args = append(a.args, "--vault", a.vault)
	}
	if len(userArgs) > 2 && userArgs[2] != "" {
		a.account = userArgs[2]
		a.args = append(a.args, "--account", a.account)
	}
	return a, nil
}
