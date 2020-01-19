package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	minlen  uint
	nodesep float64
)

// location of dot executable for converting from .dot to .svg
// it's usually at: /usr/bin/dot
var dotExe string

// dotToImage generates a SVG using the 'dot' utility, returning the filepath
func dotToImage(outfname string, format string, dot []byte) (string, error) {
	if dotExe == "" {
		dot, err := exec.LookPath("dot")
		if err != nil {
			log.Fatalln("unable to find program 'dot', please install it or check your PATH")
		}
		dotExe = dot
	}

	var img string
	if outfname == "" {
		img = filepath.Join(os.TempDir(), fmt.Sprintf("go-callvis_export.%s", format))
	} else {
		img = fmt.Sprintf("%s.%s", outfname, format)
	}
	ioutil.WriteFile("test.dot",dot, 0644)
	cmd := exec.Command(dotExe, fmt.Sprintf("-T%s", format), "-o", img)
	cmd.Stdin = bytes.NewReader(dot)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return img, nil
}

const (
	FontTitle = "Consolas"
	FontNode  = "Tahoma"
)

const tmplCluster = `{{define "cluster" -}}
    {{printf "subgraph %q {" .}}
        {{printf "%s" .Attrs.Lines}}
        {{range .Nodes}}
        {{template "node" .}}
        {{- end}}
        {{range .Clusters}}
        {{template "cluster" .}}
        {{- end}}
    {{println "}" }}
{{- end}}`

const tmplNode = `{{define "edge" -}}
    {{printf "%q -> %q [ %s ]" .From .To .Attrs}}
{{- end}}`

const tmplEdge = `{{define "node" -}}
    {{printf "%q [ %s ]" .ID .Attrs}}
{{- end}}`

const tmplGraph = `digraph gocallvis {
    label="{{.Title}}";
    labeljust="l";
    fontname="Arial";
    fontsize="14";
    rankdir="LR";
    bgcolor="lightgray";
    style="solid";
    penwidth="0.5";
    pad="0.0";
    nodesep="{{.Options.nodesep}}";

    node [shape="ellipse" style="filled" fillcolor="honeydew" fontname="Verdana" penwidth="1.0" margin="0.05,0.0"];
    edge [minlen="{{.Options.minlen}}"]

    {{template "cluster" .Cluster}}

    {{- range .Edges}}
    {{template "edge" .}}
    {{- end}}
}
`

func WriteDot(w io.Writer, g *dotGraph) error {
	t := template.New("dot")
	for _, s := range []string{tmplCluster, tmplNode, tmplEdge, tmplGraph} {
		if _, err := t.Parse(s); err != nil {
			return err
		}
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, g); err != nil {
		return err
	}
	_, err := buf.WriteTo(w)
	return err
}

type dotGraph struct {
	Title   string
	Minlen  uint
	Attrs   dotAttrs
	Cluster *dotCluster
	Nodes   []*dotNode
	Edges   []*dotEdge
	Options map[string]string
}

type dotCluster struct {
	ID       string
	Clusters map[string]*dotCluster
	Nodes    []*dotNode
	Attrs    dotAttrs
}

func NewDotCluster(id string) *dotCluster {
	return &dotCluster{
		ID:       id,
		Clusters: make(map[string]*dotCluster),
		Attrs:    make(dotAttrs),
	}
}

func (c *dotCluster) String() string {
	return fmt.Sprintf("cluster_%s", c.ID)
}

type dotNode struct {
	ID    string
	Attrs dotAttrs
}

func (n *dotNode) String() string {
	return n.ID
}

type dotEdge struct {
	From  *dotNode
	To    *dotNode
	Attrs dotAttrs
}

type dotAttrs map[string]string

func (p dotAttrs) List() []string {
	l := []string{}
	for k, v := range p {
		l = append(l, fmt.Sprintf("%s=%q", k, v))
	}
	return l
}

func (p dotAttrs) String() string {
	return strings.Join(p.List(), " ")
}

func (p dotAttrs) Lines() string {
	return fmt.Sprintf("%s;", strings.Join(p.List(), ";\n"))
}
