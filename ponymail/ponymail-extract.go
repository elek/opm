package ponymail

import (
	"bufio"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/writer"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"path"
	"regexp"
	"strings"
	"time"
)

func init() {
	cmd := cli.Command{
		Name: "extract",
		Action: func(c *cli.Context) error {
			store, err := kv.Create(c.Args().Get(0))
			if err != nil {
				return err
			}
			return extract(store)
		},
	}
	RegisterPonymail(cmd)

}

func extract(store kv.KV) error {
	w, err := writer.NewWriter("ponymail", "csv", new(Mail))
	defer w.Close()
	if err != nil {
		return err
	}
	archivePath := path.Join("ponymail", "archive")
	return store.IterateSubTree(archivePath, func(key string) error {
		return extractMonth(store, key, w)
	})
	return nil
}

type Mail struct {
	List        string
	Domain      string
	From        string
	Subject     string
	Date        time.Time
	MessageId   string
	ReplyTo     string
	RecipientNo int
}

func extractMonth(k kv.KV, key string, w writer.Writer) error {
	log.Info().Msg("Extracting mail data from " + key)
	content, err := k.GetReader(key)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(content)
	scanner.Split(bufio.ScanLines)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	header := true

	var fromPattern = regexp.MustCompile(`^From: .*<(.+)>.*$`)
	var fromPatternSimple = regexp.MustCompile(`^From: (.+)$`)

	m := Mail{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "From ") {
			header = true
			if m.Subject != "" {
				err = w.Write(m)
				if err != nil {
					return err
				}
			}
			m = Mail{
				Domain: path.Base(path.Dir(path.Dir(path.Dir(key)))),
				List:   path.Base(path.Dir(path.Dir(key))),
			}
		}
		if strings.TrimSpace(line) == "" {
			header = false
		}
		if header == true {
			if strings.HasPrefix(line, "From: ") {
				matches := fromPattern.FindSubmatch([]byte(line))
				if len(matches) == 0 {
					matches = fromPatternSimple.FindSubmatch([]byte(line))
				}
				m.From = string(matches[1])
			} else if strings.HasPrefix(line, "Subject: ") {
				m.Subject = line[9:]
			} else if strings.HasPrefix(line, "Date: ") {
				parts := strings.Split(line[6:], "(")
				t, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", strings.TrimSpace(parts[0]))
				if err != nil {
					return err
				}
				m.Date = t
			} else if strings.HasPrefix(line, "Message-ID: ") {
				m.MessageId = strings.Split(line, " ")[1]
			} else if strings.HasPrefix(line, "Reply-To: ") {
				m.ReplyTo = strings.Split(line, " ")[1]
			} else if strings.HasPrefix(line, "To: ") {
				m.RecipientNo = len(strings.Split(line, ","))
			}
		}
	}
	if m.Subject != "" {
		err = w.Write(m)
		if err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
