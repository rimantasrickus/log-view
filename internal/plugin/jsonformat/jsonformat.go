package jsonformat

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"log-viewer/internal/plugin"
)

func init() {
	plugin.Register(&JSONFormatter{})
}

// JSONFormatter is a plugin that formats JSON log lines into indented, readable JSON.
type JSONFormatter struct{}

func (j *JSONFormatter) Name() string {
	return "Format as JSON"
}

func (j *JSONFormatter) Description() string {
	return "Pretty-print JSON content in a formatted view"
}

func (j *JSONFormatter) CanHandle(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 2 {
		return false
	}
	if trimmed[0] != '{' || trimmed[len(trimmed)-1] != '}' {
		return false
	}
	return json.Valid([]byte(trimmed))
}

func (j *JSONFormatter) Render(line string) fyne.CanvasObject {
	trimmed := strings.TrimSpace(line)

	var raw any
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		entry := widget.NewMultiLineEntry()
		entry.SetText(line)
		entry.Disable()
		return entry
	}

	expanded := expandJSONStrings(raw)

	// Build a tree model from the parsed JSON
	model := newJSONTreeModel(expanded)

	var tree *widget.Tree
	tree = widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			return model.childUIDs(uid)
		},
		func(uid widget.TreeNodeID) bool {
			return model.isBranch(uid)
		},
		func(branch bool) fyne.CanvasObject {
			txt := canvas.NewText("", nil)
			txt.TextStyle = fyne.TextStyle{Monospace: true}
			txt.TextSize = 12
			return txt
		},
		func(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			txt := obj.(*canvas.Text)
			txt.Text = model.nodeLabel(uid, branch && tree.IsBranchOpen(uid), tree)
			txt.Refresh()
		},
	)
	model.tree = tree

	// Open all branches by default
	model.openAll(tree, "")

	return container.NewStack(tree)
}

// expandJSONStrings recursively walks a parsed JSON value and expands any
// string values that are themselves valid JSON objects or arrays.
func expandJSONStrings(v any) any {
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			val[k] = expandJSONStrings(child)
		}
		return val
	case []any:
		for i, child := range val {
			val[i] = expandJSONStrings(child)
		}
		return val
	case string:
		trimmed := strings.TrimSpace(val)
		if len(trimmed) >= 2 && (trimmed[0] == '{' || trimmed[0] == '[') {
			var parsed any
			if json.Unmarshal([]byte(trimmed), &parsed) == nil {
				return expandJSONStrings(parsed)
			}
		}
		return val
	default:
		return val
	}
}

// jsonTreeModel holds the tree structure for the JSON tree widget.
type jsonTreeModel struct {
	nodes map[string]*jsonNode
	tree  *widget.Tree
}

type jsonNode struct {
	key      string // display key (empty for root)
	value    any    // the raw value (for leaves)
	children []string
	isBranch bool
	nodeType string // "object", "array", "leaf"
	count    int    // number of children (for collapsed label)
}

func newJSONTreeModel(data any) *jsonTreeModel {
	m := &jsonTreeModel{
		nodes: make(map[string]*jsonNode),
	}
	m.buildNode("", "", data)
	return m
}

func (m *jsonTreeModel) openAll(tree *widget.Tree, uid widget.TreeNodeID) {
	node, ok := m.nodes[uid]
	if !ok {
		return
	}
	if node.isBranch {
		tree.OpenBranch(uid)
		for _, child := range node.children {
			m.openAll(tree, child)
		}
	}
}

func (m *jsonTreeModel) buildNode(uid, key string, value any) {
	switch val := value.(type) {
	case map[string]any:
		node := &jsonNode{
			key:      key,
			isBranch: true,
			nodeType: "object",
			count:    len(val),
		}
		// Sort keys for stable ordering
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			childUID := uid + "/" + k
			node.children = append(node.children, childUID)
			m.buildNode(childUID, k, val[k])
			// If child is a branch, add its closing bracket as sibling
			if childNode, ok := m.nodes[childUID]; ok && childNode.isBranch {
				closeUID := childUID + "/__close"
				bracket := "}"
				if childNode.nodeType == "array" {
					bracket = "]"
				}
				node.children = append(node.children, closeUID)
				m.nodes[closeUID] = &jsonNode{key: bracket, nodeType: "close"}
			}
		}
		m.nodes[uid] = node

	case []any:
		node := &jsonNode{
			key:      key,
			isBranch: true,
			nodeType: "array",
			count:    len(val),
		}
		for i, child := range val {
			childUID := fmt.Sprintf("%s/[%d]", uid, i)
			node.children = append(node.children, childUID)
			m.buildNode(childUID, fmt.Sprintf("[%d]", i), child)
			// If child is a branch, add its closing bracket as sibling
			if childNode, ok := m.nodes[childUID]; ok && childNode.isBranch {
				closeUID := childUID + "/__close"
				bracket := "}"
				if childNode.nodeType == "array" {
					bracket = "]"
				}
				node.children = append(node.children, closeUID)
				m.nodes[closeUID] = &jsonNode{key: bracket, nodeType: "close"}
			}
		}
		m.nodes[uid] = node

	default:
		node := &jsonNode{
			key:      key,
			value:    val,
			isBranch: false,
			nodeType: "leaf",
		}
		m.nodes[uid] = node
	}
}

func (m *jsonTreeModel) childUIDs(uid widget.TreeNodeID) []widget.TreeNodeID {
	node, ok := m.nodes[uid]
	if !ok {
		return nil
	}
	if m.tree == nil {
		return node.children
	}
	// Filter out close bracket nodes whose branch is closed
	result := make([]widget.TreeNodeID, 0, len(node.children))
	for _, child := range node.children {
		if strings.HasSuffix(child, "/__close") {
			branchUID := strings.TrimSuffix(child, "/__close")
			if !m.tree.IsBranchOpen(branchUID) {
				continue
			}
		}
		result = append(result, child)
	}
	return result
}

func (m *jsonTreeModel) isBranch(uid widget.TreeNodeID) bool {
	node, ok := m.nodes[uid]
	if !ok {
		return false
	}
	return node.isBranch
}

func (m *jsonTreeModel) nodeText(uid widget.TreeNodeID) string {
	return m.nodeLabel(uid, false, nil)
}

func (m *jsonTreeModel) nodeLabel(uid widget.TreeNodeID, open bool, tree *widget.Tree) string {
	node, ok := m.nodes[uid]
	if !ok {
		return ""
	}

	// Closing bracket nodes — only show when associated branch is open
	if node.nodeType == "close" {
		if tree != nil && strings.HasSuffix(string(uid), "/__close") {
			branchUID := strings.TrimSuffix(string(uid), "/__close")
			if !tree.IsBranchOpen(branchUID) {
				return ""
			}
		}
		return node.key
	}

	var sb strings.Builder

	if node.key != "" {
		sb.WriteString(node.key)
	}

	if node.isBranch {
		if open {
			if node.nodeType == "object" {
				if node.key != "" {
					sb.WriteString(": {")
				} else {
					sb.WriteString("{")
				}
			} else {
				if node.key != "" {
					sb.WriteString(": [")
				} else {
					sb.WriteString("[")
				}
			}
		} else {
			if node.key != "" {
				sb.WriteString(": ")
			}
			if node.nodeType == "object" {
				sb.WriteString(fmt.Sprintf("{...} %d keys", node.count))
			} else {
				sb.WriteString(fmt.Sprintf("[...] %d items", node.count))
			}
		}
	} else {
		if node.key != "" {
			sb.WriteString(": ")
		}
		sb.WriteString(formatValue(node.value))
	}

	return sb.String()
}

func formatValue(val any) string {
	switch v := val.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}
