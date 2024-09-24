package Common

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
)

type FileTreeNode struct {
	Children map[string]*FileTreeNode
	Name     string
	Selected bool
}

func (f *FileTreeNode) GenerateTree(t *tree.Tree) {
	var prefix string

	// iterate over the map in sorted order
	keys := make([]string, 0, len(f.Children))
	for key := range f.Children {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		child := f.Children[key]
		if strings.HasSuffix(key, ".mq4") {
			prefix = " "
		} else {
			prefix = " "
		}

		fileName := prefix + key

		if child.Selected {
			key = lipgloss.JoinHorizontal(
				lipgloss.Left,
				lipgloss.NewStyle().Foreground(lipgloss.Color("#2f334d")).Render(LEFT_HALF_CIRCLE),
				lipgloss.NewStyle().Background(lipgloss.Color("#2f334d")).Render(key),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#2f334d")).Render(RIGHT_HALF_CIRCLE),
			)
			fileName = lipgloss.
				NewStyle().
				Foreground(lipgloss.Color("#10a3be")).
				Render(prefix) + key
		}

		newLeaf := tree.Root(fileName)
		t.Child(newLeaf)
		child.GenerateTree(newLeaf)
	}
}

type File struct {
	Path     string
	Content  string
	Selected bool
}

func getFiles(root string) []File {
	files := []File{}

	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".mq4") {
			return nil
		}

		files = append(files, File{
			Path: path,
		})
		return nil
	})

	// set the first file as selected
	files[0].Selected = true

	return files
}

func (m FilePicker) buildTree(isMain bool) string {
	enumeratorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingRight(1)

	pwd, _ := os.Getwd()

	folderName := filepath.Base(pwd)

	t := tree.New().
		Root(folderName).
		RootStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("111"))).
		Enumerator(tree.RoundedEnumerator).
		EnumeratorStyle(enumeratorStyle)

	fileTree := FileTreeNode{
		Name:     folderName,
		Children: make(map[string]*FileTreeNode),
	}

	for _, file := range m.Files {
		chunks := strings.Split(file.Path, string(os.PathSeparator))

		currNode := &fileTree

		for i := 0; i < len(chunks); i++ {
			// check if the folder already exists
			if _, ok := currNode.Children[chunks[i]]; ok {
				currNode = currNode.Children[chunks[i]]
				continue
			}
			currNode.Children[chunks[i]] = &FileTreeNode{
				Name:     chunks[i],
				Children: make(map[string]*FileTreeNode),
				Selected: file.Selected && i == len(chunks)-1,
			}
			currNode = currNode.Children[chunks[i]]
		}
	}

	fileTree.GenerateTree(t)

	t.ItemStyleFunc(func(d tree.Children, i int) lipgloss.Style {
		if d.At(i).Children().Length() == 0 {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#828bb8"))
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	})

	if isMain {
		subTreeState := m

		subTreeState.Files = m.Files[:m.CurrIndex]

		subTree := subTreeState.buildTree(false)

		height := lipgloss.Height(subTree)

		// return t.String with an offset
		treeStr := t.String()
		treeStrSplit := strings.Split(treeStr, "\n")
		if height > 0 {
			height--
		}

		// check if it even overflows
		if len(treeStrSplit) < m.height {
			return treeStr
		}

		if height < m.height/2 {
			height = 0
		}

		end := height + m.height
		if end > len(treeStrSplit) {
			end = len(treeStrSplit)
		}

		// adjust the start if the tree can fit in the current window
		if end-height < m.height {
			height = end - m.height
		}
		return strings.Join(treeStrSplit[height:end], "\n")
	}
	return t.String()
}
