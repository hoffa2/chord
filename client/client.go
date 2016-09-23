package client

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	// IP addresses to chord nodes
	IPs []string
	// Ip address of nameserver
	nameServer string
	// Connection object - used for all http interaction
	conn      http.Client
	results   chan time.Duration
	keyvalues map[string]string
	nkeys     int
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const keysize = 160

func (c *Client) putKey(key, val, nodeip string) error {
	url := fmt.Sprintf("http://%s/%s", nodeip+":8030", key)
	req, err := http.NewRequest("PUT", url, strings.NewReader(val))
	if err != nil {
		return err
	}
	before := time.Now()
	resp, err := c.conn.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	c.results <- time.Now().Sub(before)

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Unsuccesful PUT request (%s)\tErr: %s", string(body))
	}
	io.Copy(ioutil.Discard, resp.Body)
	return nil
}

func (c *Client) getKey(key, nodeip string) error {
	url := fmt.Sprintf("http://%s/%s", nodeip+":8030", key)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	before := time.Now()
	resp, err := c.conn.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	c.results <- time.Now().Sub(before)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unsuccesful GET request (%s)\tErr: %s", key, string(body))
	}
	if strings.Compare(string(body), c.keyvalues[key]) != 0 {
		return fmt.Errorf("Get returned wrong key (%s:%s)", string(body), c.keyvalues[key])
	}

	return nil
}

func (c *Client) RunTests() error {
	for i := 0; i < c.nkeys; i++ {
		c.keyvalues[strconv.Itoa(i*100)] = strconv.Itoa(i * 100)
	}

	start := time.Now()
	for k, v := range c.keyvalues {
		err := c.putKey(k, v, c.IPs[rand.Intn(len(c.IPs))])
		if err != nil {
			return err
		}
	}

	for k, _ := range c.keyvalues {
		err := c.getKey(k, c.IPs[rand.Intn(len(c.IPs))])
		if err != nil {
			return err
		}
	}
	end := time.Now().Sub(start)
	c.finalize(end)
	return nil
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (c *Client) finalize(totalTime time.Duration) {
	var times []float64
	var avgTotal float64
	for {
		select {
		case r := <-c.results:
			times = append(times, r.Seconds())
			avgTotal += r.Seconds()
		default:
			fmt.Printf("\tTotal:\t%4.4f secs\n", totalTime.Seconds())
			avgTime := avgTotal / float64(len(times))
			fmt.Printf("\tRequests/s\t%4.4f\n", float64(len(times))/totalTime.Seconds())
			fmt.Printf("\tMeanLatenct:\t%4.4f\n", avgTime)
			return
		}
	}
}
