package md

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path"

	"metascoop/apps"
)

const (
	tableStart = "<!-- This table is auto-generated. Do not edit -->"

	tableEnd = "<!-- end apps table -->"

	tableTmpl = `
| Icon | Name | Description | Version |
| --- | --- | --- | --- |{{range .}}
| {{if .Icon}}<a href="{{.SourceCode}}"><img src="{{.Icon}}" alt="{{.Name}} icon" width="36px" height="36px"></a>{{end}} | [**{{.Name}}**]({{.SourceCode}}) | {{.Summary}} | {{.Version}} ({{.VersionCode}}) |{{end}}
` + tableEnd
)

var tmpl = template.Must(template.New("").Parse(tableTmpl))

type appRow struct {
	Name        string
	SourceCode  string
	Summary     string
	Version     string
	VersionCode string
	Icon        string
}

func appRows(index *apps.RepoIndex) []appRow {
	var rows []appRow

	for _, app := range index.Apps {
		// Files sitting in repo/ that are not APKs are indexed as pseudo-apps
		// with no source of their own; they are not something to advertise.
		sourceCode := indexString(app, "sourceCode")
		if sourceCode == "" {
			continue
		}

		rows = append(rows, appRow{
			Name:        indexString(app, "name"),
			SourceCode:  sourceCode,
			Summary:     indexString(app, "summary"),
			Version:     indexString(app, "suggestedVersionName"),
			VersionCode: indexString(app, "suggestedVersionCode"),
			Icon:        iconPath(app),
		})
	}

	return rows
}

// iconPath prefers the icon supplied through metadata, which is where an app
// whose launcher icon is adaptive (and so has no raster for fdroid to pull out
// of the APK) gets one.
func iconPath(app map[string]interface{}) string {
	packageName := indexString(app, "packageName")

	if localized, ok := app["localized"].(map[string]interface{}); ok {
		if enUS, ok := localized["en-US"].(map[string]interface{}); ok {
			if icon := indexString(enUS, "icon"); icon != "" && packageName != "" {
				return path.Join("fdroid/repo", packageName, "en-US", icon)
			}
		}
	}

	if icon := indexString(app, "icon"); icon != "" {
		return path.Join("fdroid/repo/icons", icon)
	}

	return ""
}

func indexString(m map[string]interface{}, key string) string {
	s, _ := m[key].(string)
	return s
}

func RegenerateReadme(readMePath string, index *apps.RepoIndex) (err error) {
	content, err := os.ReadFile(readMePath)
	if err != nil {
		return
	}

	var tableStartIndex = bytes.Index(content, []byte(tableStart))
	if tableStartIndex < 0 {
		return fmt.Errorf("cannot find table start in %q", readMePath)
	}

	var tableEndIndex = bytes.Index(content, []byte(tableEnd))
	if tableEndIndex < 0 {
		return fmt.Errorf("cannot find table end in %q", readMePath)
	}

	var table bytes.Buffer

	table.WriteString(tableStart)

	err = tmpl.Execute(&table, appRows(index))
	if err != nil {
		return err
	}

	newContent := []byte{}

	newContent = append(newContent, content[:tableStartIndex]...)
	newContent = append(newContent, table.Bytes()...)
	newContent = append(newContent, content[tableEndIndex:]...)

	return os.WriteFile(readMePath, newContent, os.ModePerm)
}
