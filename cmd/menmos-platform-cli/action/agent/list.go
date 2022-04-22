package agent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/menmos/menmos-agent/payload"
	"github.com/urfave/cli/v2"
)

const LIST_AGENT_PATH = "node"

func List(c *cli.Context) error {
	agent := c.String("host")
	u, err := url.Parse(agent)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, LIST_AGENT_PATH)

	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var body payload.ListNodesResponse
	if err := json.Unmarshal(raw, &body); err != nil {
		return err
	}

	for _, node := range body.Nodes {
		fmt.Println(node.ID)
	}

	return nil
}
