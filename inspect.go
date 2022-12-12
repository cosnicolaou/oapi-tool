package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/cosnicolaou/openapi"
	"github.com/getkin/kin-openapi/openapi3"
)

type InspectFlags struct {
	Path       string `subcmd:"path,,colon separated path of the component in the spec to inspect"`
	FollowRegs bool   `subcmd:"follow-refs,false,set to true to follow/flatter schema references"`
	Recurse    bool   `subcmd:"recurse,false,recurse into all sub-nodes of the requested path"`
	TracePaths bool   `subcmd:"trace-paths,false,display each path in the spec as it is visited"`
}

func inspectCmd(ctx context.Context, values any, args []string) error {
	fv := values.(*InspectFlags)
	filename := args[0]
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(filename)
	if err != nil {
		return err
	}
	visitor := &visitor{
		recurse: fv.Recurse,
	}

	opts := []openapi.WalkerOption{openapi.WalkerTracePaths(fv.TracePaths)}
	if len(fv.Path) > 0 {
		opts = append(opts, openapi.WalkerVisitPrefix(strings.Split(fv.Path, ":")...))
	}
	walker := openapi.NewWalker(visitor.visit, opts...)
	walker.Walk(doc)
	return nil
}

type visitor struct {
	recurse bool
}

func indent(path []string, indent int, node any) (string, error) {
	out := strings.Builder{}
	for i, p := range path {
		out.WriteString(strings.Repeat(" ", i*indent))
		out.WriteString(p)
		out.WriteString(":\n")
	}
	lines, err := openapi.AsYAML(indent, node)
	if err != nil {
		return "", err
	}
	prefix := strings.Repeat(" ", len(path)*indent)
	for _, l := range strings.Split(lines, "\n") {
		out.WriteString(prefix)
		out.WriteString(l)
		out.WriteRune('\n')
	}
	return out.String(), nil
}

func (v visitor) visit(path []string, parent, node any) (bool, error) {
	buf, err := indent(path, 2, node)
	if err != nil {
		fmt.Printf("%v: %v\n", strings.Join(path, ":"), err)
		return v.recurse, nil

	}
	fmt.Println(buf)
	return v.recurse, nil
}
