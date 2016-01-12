package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	volumesPath      = "/api/1.0/volumes"
	volumeCreatePath = "/api/1.0/volume/%s"
	volumeStopPath   = "/api/1.0/volume/%s/stop"
)

type peer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type volume struct {
	Name       string `json:"name"`
	UUID       string `json:"uuid"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	NumBricks  int    `json:"num_bricks"`
	Distribute int    `json:"distribute"`
	Stripe     int    `json:"stripe"`
	Replica    int    `json:"replica"`
	Transport  string `json:"transport"`
}

type response struct {
	Ok  bool   `json:"ok"`
	Err string `json:"error,omitempty"`
}

type peerResponse struct {
	Data []peer `json:"data,omitempty"`
	response
}

type volumeResponse struct {
	Data []volume `json:"data,omitempty"`
	response
}

// Client is the http client that sends requests to the gluster API.
type Client struct {
	addr string
	base string
}

// NewClient initializes a new client.
func NewClient(addr, base string) *Client {
	return &Client{addr, base}
}

// VolumeExist returns whether a volume exist in the cluster with a given name or not.
func (r Client) VolumeExist(name string) (bool, error) {
	vols, err := r.volumes()
	if err != nil {
		return false, err
	}

	for _, v := range vols {
		if v.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (r Client) volumes() ([]volume, error) {
	u := fmt.Sprintf("%s%s", r.addr, volumesPath)

	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	var d volumeResponse
	if err := json.NewDecoder(res.Body).Decode(&d); err != nil {
		return nil, err
	}

	if !d.Ok {
		return nil, fmt.Errorf(d.Err)
	}
	return d.Data, nil
}

// CreateVolume creates a new volume with the given name in the cluster.
func (r Client) CreateVolume(name string, peers []string) error {
	u := fmt.Sprintf("%s%s", r.addr, fmt.Sprintf(volumeCreatePath, name))
	fmt.Println(u)

	bricks := make([]string, len(peers))
	for i, p := range peers {
		bricks[i] = fmt.Sprintf("%s:%s", p, filepath.Join(r.base, name))
	}

	params := url.Values{
		"bricks":    {strings.Join(bricks, ",")},
		"replica":   {strconv.Itoa(len(peers))},
		"transport": {"tcp"},
		"start":     {"true"},
		"force":     {"true"},
	}

	resp, err := http.PostForm(u, params)
	if err != nil {
		return err
	}

	return responseCheck(resp)
}

// StopVolume stops the volume with the given name in the cluster.
func (r Client) StopVolume(name string) error {
	u := fmt.Sprintf("%s%s", r.addr, fmt.Sprintf(volumeStopPath, name))

	req, err := http.NewRequest("PUT", u, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	return responseCheck(resp)
}

func responseCheck(resp *http.Response) error {
	var p response
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return err
	}

	if !p.Ok {
		return fmt.Errorf(p.Err)
	}

	return nil
}
