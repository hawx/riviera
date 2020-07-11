package riverjs

import (
	"html/template"
	"time"
)

func Time(t time.Time) RssTime {
	return RssTime{t}
}

// RssTime wraps a time.Time object so that when serialised and unserialised it
// uses the RFC1123Z format.
type RssTime struct {
	time.Time
}

func (t RssTime) MarshalText() ([]byte, error) {
	return []byte(t.Format(time.RFC1123Z)), nil
}

func (t RssTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Format(time.RFC1123Z) + `"`), nil
}

func (t *RssTime) UnmarshalText(data []byte) error {
	g, err := time.Parse(time.RFC1123Z, string(data))
	if err != nil {
		return err
	}
	*t = RssTime{g}
	return nil
}

func (t *RssTime) UnmarshalJSON(data []byte) error {
	g, err := time.Parse(`"`+time.RFC1123Z+`"`, string(data))
	if err != nil {
		return err
	}
	*t = RssTime{g}
	return nil
}

func (t *RssTime) HtmlFormat() template.HTML {
	return template.HTML("<time pubdate=\"" + t.Format(time.RFC3339) + "\">" +
		t.Format("02 Jan; 15:04 PM") +
		"</time>")
}
