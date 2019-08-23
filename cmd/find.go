package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
	"github.com/urfave/cli"
)

var Find = cli.Command{
	Name:  "find",
	Usage: "Find break point",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "cap, c",
			Usage: "After how many `MILLISECOND` is it considered a fail. it looks at the 95th percentil",
			Value: 500,
		},
		cli.IntFlag{
			Name:  "rate, r",
			Usage: "the start rate (`RPS`)",
			Value: 4,
		},
		cli.IntFlag{
			Name:  "duration, d",
			Usage: "duration is `SEC` of each attacks",
			Value: 15,
		},
	},
	Action: find,
}

func find(c *cli.Context) error {
	rate := c.Int("rate")
	okRate := 1
	var nokRate int
	sla := time.Duration(c.Int("cap")) * time.Millisecond
	fmt.Println(sla)

	file, err := os.Open(c.GlobalString("target"))
	if err != nil {
		return err
	}

	h := http.Header{}

	if c.GlobalString("auth") != "" {
		h["Authorization"] = []string{fmt.Sprintf("basic %s", c.GlobalString("auth"))}
	}

	if c.GlobalBool("nocache") {
		h["Cache-Control"] = []string{"no-cache"}
	}

	d := c.Int("duration")
	reader := bufio.NewReader(file)
	targets, _ := vegeta.NewEagerTargeter(reader, []byte{}, h)
	// first, find the point at which the system breaks
	for {
		if testRate(rate, sla, targets, d) {
			okRate = rate
			rate *= 2
		} else {
			nokRate = rate
			break
		}
	}

	// next, do a binary search between okRate and nokRate
	for (nokRate - okRate) > 1 {
		rate = (nokRate + okRate) / 2
		if testRate(rate, sla, targets, d) {
			okRate = rate
		} else {
			nokRate = rate
		}
	}
	fmt.Printf("➡️  Maximum Working Rate: %d req/sec\n", okRate)
	return nil
}

func testRate(rate int, sla time.Duration, t vegeta.Targeter, d int) bool {
	duration := time.Duration(d) * time.Second
	attacker := vegeta.NewAttacker()
	var metrics vegeta.Metrics
	for res := range attacker.Attack(t, uint64(rate), duration, "") {
		metrics.Add(res)
	}
	metrics.Close()
	latency := metrics.Latencies.P95
	codes, _ := json.MarshalIndent(metrics.StatusCodes, "", " ")
	if latency > sla {
		fmt.Printf("😡; Failed at %d req/sec (latency %s) --- responses\n%s\n", rate, latency, string(codes))
		return false
	}

	for k, _ := range metrics.StatusCodes {
		c, err := strconv.Atoi(k)
		if err != nil {
			panic(err)
		}
		if c >= 500 || c == 0 {
			fmt.Printf("😡; Failed at %d req/sec (latency %s) 0 codes --- responses\n%s\n", rate, latency, string(codes))
			return false
		}
	}

	fmt.Printf("😃; Success at %d req/sec (latency %s) --- responses:\n%s\n", rate, latency, string(codes))
	if len(metrics.Errors) > 0 {
		fmt.Println("with ERRORS:")
		for _, v := range metrics.Errors {
			fmt.Println(v)
		}
	}
	return true
}
