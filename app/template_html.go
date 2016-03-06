package app

import (
	"bytes"
	"html/template"
)

var (
	funcMap = template.FuncMap{
		"tripRowStyle": func(trip *Trip) string {
			switch trip.Update {
			case FareIncrease:
				return "background-color: #F79F79"
			case FareDecrease:
				return "background-color: #E3F09B"
			case Added:
				return "background-color: #87B6A7"
			case Removed:
				return "background-color: #F79F79"
			default:
				return ""
			}
		},
	}

  htmlTemplate = template.Must(template.New("HtmlBody").Funcs(funcMap).Parse(htmlTemplateDef))
)

func BodyToHTML(body *Body) (string, error) {
	var buf bytes.Buffer
	if err := htmlTemplate.Execute(&buf, body); err != nil {
		return "", err
	}
	return buf.String(), nil
}

var htmlTemplateDef = `
<html>
  <head></head>
  <body style="font-family: Arial, Helvetica, sans-serif">
    {{range .SearchGroups}}
      <div>
        <h3 style="margin-bottom: 3px;">{{.Date}}</h3>
        {{if .MaxFare}}<i>Max {{.MaxFare}}</i>{{end}}
        {{if .Note}}<i>({{.Note}})</i>{{end}}
        <table style="border-collapse: collapse">
          <thead>
            <th colspan="2" style="text-align: left; padding: 0px 15px 0px 0px">From</th>
            <th colspan="2" style="text-align: left; padding: 0px 15px 0px 0px">To</th>
            <th style="text-align: left; padding: 0px 15px 0px 0px">Stops</th>
            <th style="text-align: left; padding: 0px 15px 0px 0px">Price</th>
          </thead>
          <tbody>
            {{range .Trips}}
              <tr style="{{. | tripRowStyle}}">
                <td style="padding: 0px 15px 0px 0px">{{.OriginAirport}}</td>
                <td style="padding: 0px 15px 0px 0px">{{.DepartureTime}}</td>
                <td style="padding: 0px 15px 0px 0px">{{.DestinationAirport}}</td>
                <td style="padding: 0px 15px 0px 0px">{{.ArrivalTime}}</td>
                <td style="padding: 0px 15px 0px 0px">{{.Stops}}</td>
                <td style="padding: 0px 15px 0px 0px">{{.Cost}}</td>
              </tr>
            {{end}}
          </tbody>
      </div>
    {{end}}
  </body>
</html>
`
