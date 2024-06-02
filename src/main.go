package main

import (
	"fmt"
	"os"

	"golang.org/x/net/html"
)

// check compares two html nodes and returns true if they are different
// Exits on first difference found
func check(a, b *html.Node, head bool) bool {
	if a == nil && b == nil {
		return false
	}

	if !head &&
		a.Type == html.ElementNode &&
		b.Type == html.ElementNode &&
		a.Data == "head" &&
		b.Data == "head" {
		return check(a.NextSibling, b.NextSibling, head)
	}

	// TODO: Check attributes
	// if n.Data == "a" {
	//     for _, a := range n.Attr {
	//         fmt.Println(a.Key, a.Val)
	//     }
	// }

	if a.Data != b.Data {
		return true
	}

	isDifferent := false

	isDifferent = check(a.FirstChild, b.FirstChild, head)
	if isDifferent {
		return true
	}

	isDifferent = check(a.NextSibling, b.NextSibling, head)

	return isDifferent
}

func makeChangeWrapper(node *html.Node, color string) *html.Node {
    newNode := &html.Node{
        Type:      node.Type,
        DataAtom:  node.DataAtom,
        Data:      node.Data,
        Namespace: node.Namespace,
        Attr:      node.Attr,
    }

    styleWrapper := &html.Node{
        Type: html.ElementNode,
        Data: "span",
        Attr: []html.Attribute{
            {
                Key: "style",
                Val: fmt.Sprintf("background-color: %s;", color),
            },
        },
        FirstChild: newNode,
    }

    return styleWrapper
}


// createDiff compares two html nodes and edits the first one to show the differences
func createDiff(a, b, parent *html.Node) {
	if a == nil && b == nil {
		return
	}

	if a == nil {
		parent.AppendChild(makeChangeWrapper(b, "green"))
		return
	}

	if b == nil {
        a.Parent.AppendChild(
            makeChangeWrapper(a.FirstChild, "red"),
        )

        a.Parent.RemoveChild(a)
		return
	}

	// Skip head
	if a.Type == html.ElementNode &&
		b.Type == html.ElementNode &&
		a.Data == "head" &&
		b.Data == "head" {
		createDiff(a.NextSibling, b.NextSibling, a)
		return
	}

	if a.Data != b.Data {
        newNode := makeChangeWrapper(a, "green")

		parent.InsertBefore(
            newNode,
			a,
		)

        parent.InsertBefore(
            makeChangeWrapper(b, "red"),
            newNode,
        )
    
        defer a.Parent.RemoveChild(a)
	}

	createDiff(a.FirstChild, b.FirstChild, a)
	createDiff(a.NextSibling, b.NextSibling, a)
}

func openFiles(fileA, fileB string) (*html.Node, *html.Node) {
	fA, err := os.Open(fileA)
	if err != nil {
		panic(fmt.Sprintf("Could not open html file: %s", err))
	}
	defer fA.Close()

	docA, err := html.Parse(fA)
	if err != nil {
		panic(fmt.Sprintf("parsing html: %s", err))
	}

	fB, err := os.Open(fileB)
	if err != nil {
		panic(fmt.Sprintf("Could not read html file: %s", err))
	}
	defer fB.Close()

	docB, err := html.Parse(fB)
	if err != nil {
		panic(fmt.Sprintf("parsing html: %s", err))
	}

	return docA, docB
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: diffhtml fileA fileB")
		os.Exit(1)
	}

	fileA := os.Args[1]
	fileB := os.Args[2]

	docA, docB := openFiles(fileA, fileB)

	fmt.Println("Are files different?:", check(docA, docB, true))
	createDiff(docA, docB, docA)

	diff, err := os.Create("/tmp/diff.html")
	if err != nil {
		panic(fmt.Sprintf("Could not open diff file: %s", err))
	}
	defer diff.Close()

	err = html.Render(diff, docA)
	if err != nil {
		panic(fmt.Sprintf("Could not render diff file: %s", err))
	}
}
