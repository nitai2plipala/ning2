package spider

import (
	"strings"
	"testing"
)

// ==================== ParseHtml 基础测试 ====================

func TestParseHtml_Basic(t *testing.T) {
	html := `<html><body><div id="main">Hello</div></body></html>`
	reader := strings.NewReader(html)

	root, err := ParseHtml(reader)
	if err != nil {
		t.Errorf("ParseHtml failed: %v", err)
	}

	if root == nil {
		t.Fatal("Root node should not be nil")
	}
}

func TestParseHtml_Empty(t *testing.T) {
	_, err := ParseHtml(strings.NewReader(""))
	if err != nil {
		t.Errorf("Empty HTML should not cause error, got: %v", err)
	}
}

func TestParseHtml_OnlyText(t *testing.T) {
	root, err := ParseHtml(strings.NewReader("just some text"))
	if err != nil {
		t.Errorf("ParseHtml failed: %v", err)
	}
	if root == nil {
		t.Error("Should parse text-only HTML")
	}
}

func TestParseHtml_Invalid(t *testing.T) {
	html := `<html><body><div>unclosed`
	reader := strings.NewReader(html)

	// HTML 解析器比较宽容，可能不会报错
	root, err := ParseHtml(reader)
	if err != nil {
		t.Logf("Parser returned error (acceptable): %v", err)
	}
	// 即使有错误，root 可能仍然部分解析
	if root == nil {
		t.Error("Root should not be nil even for invalid HTML")
	}
}

// ==================== QueryNode 标签查询测试 ====================

func TestQueryNode_Tag(t *testing.T) {
	html := `<html><body><div><p>para1</p><p>para2</p></div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	// 查询所有 p 标签
	nodes := root.QueryNode([]string{"p"})
	if len(nodes) != 2 {
		t.Errorf("Expected 2 p nodes, got %d", len(nodes))
	}
}

func TestQueryNode_Tag_NotFound(t *testing.T) {
	html := `<html><body><div>test</div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	nodes := root.QueryNode([]string{"span"})
	if len(nodes) != 0 {
		t.Error("Non-matching query should return empty")
	}
}

func TestQueryNode_EmptyPattern(t *testing.T) {
	html := `<html><body><div>test</div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	nodes := root.QueryNode([]string{})
	if nodes != nil {
		t.Error("Empty pattern should return nil")
	}
}

// ==================== QueryNode ID 查询测试 ====================

func TestQueryNode_ID(t *testing.T) {
	html := `<html><body><div id="main">Content</div><div id="other">Other</div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	// 查询 id="main"
	nodes := root.QueryNode([]string{"#main"})
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node with id=main, got %d", len(nodes))
	}
	if nodes[0].Data != "div" {
		t.Errorf("Expected div, got %s", nodes[0].Data)
	}
}

func TestQueryNode_ID_NotFound(t *testing.T) {
	html := `<html><body><div id="main">Content</div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	nodes := root.QueryNode([]string{"#notexist"})
	if len(nodes) != 0 {
		t.Error("Non-existent ID should return empty")
	}
}

// ==================== QueryNode Class 查询测试 ====================

func TestQueryNode_Class(t *testing.T) {
	html := `<html><body><div class="item">1</div><div class="item">2</div><div>3</div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	// 查询 class="item"
	nodes := root.QueryNode([]string{".item"})
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes with class=item, got %d", len(nodes))
	}
}

func TestQueryNode_Class_Multiple(t *testing.T) {
	html := `<html><body><div class="foo bar baz">text</div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	// 多个 class 查询
	nodes := root.QueryNode([]string{".foo", ".bar"})
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	}
}

func TestQueryNode_Class_NotFound(t *testing.T) {
	html := `<html><body><div class="item">text</div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	nodes := root.QueryNode([]string{".notexist"})
	if len(nodes) != 0 {
		t.Error("Non-existent class should return empty")
	}
}

// ==================== QueryNode 属性查询测试 ====================

func TestQueryNode_Attribute(t *testing.T) {
	html := `<html><body><a href="link1">1</a><a href="link2">2</a><a>3</a></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	// 查询有 href 属性的元素
	nodes := root.QueryNode([]string{"[href]"})
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes with href, got %d", len(nodes))
	}
}

func TestQueryNode_AttributeValue(t *testing.T) {
	html := `<html><body><input type="text"><input type="checkbox"></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	// 查询 type="text"
	nodes := root.QueryNode([]string{"[type=text]"})
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node with type=text, got %d", len(nodes))
	}
}

func TestQueryNode_Attribute_NotFound(t *testing.T) {
	html := `<html><body><input type="text"></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	nodes := root.QueryNode([]string{"[data-not-exist]"})
	if len(nodes) != 0 {
		t.Error("Non-existent attribute should return empty")
	}
}

// ==================== QueryNode 子代查询测试 ====================

func TestQueryNode_Child(t *testing.T) {
	html := `<html><body><div><span>child</span></div><div>text</div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	// 查询 div 的直接子元素 span
	nodes := root.QueryNode([]string{"div", ">", "span"})
	if len(nodes) != 1 {
		t.Errorf("Expected 1 child span, got %d", len(nodes))
	}
}

func TestQueryNode_Descendant(t *testing.T) {
	html := `<html><body><div><p><span>deep</span></p></div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	// 查询 div 内的所有 span（后代）
	nodes := root.QueryNode([]string{"div", "span"})
	if len(nodes) != 1 {
		t.Errorf("Expected 1 descendant span, got %d", len(nodes))
	}
}

func TestQueryNode_Nested(t *testing.T) {
	html := `<html><body><div class="outer"><div class="inner"><span>text</span></div></div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	// 嵌套查询
	nodes := root.QueryNode([]string{".outer", ".inner", "span"})
	if len(nodes) != 1 {
		t.Errorf("Expected 1 span, got %d", len(nodes))
	}
}

// ==================== QueryNode 组合查询测试 ====================

func TestQueryNode_Combined(t *testing.T) {
	html := `<html><body><div class="container"><a href="url" class="link">text</a></div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	// 组合查询：div.container 下的 a.link
	nodes := root.QueryNode([]string{"div", ".container", "a", ".link"})
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	}
}

func TestQueryNode_Combined_ID_And_Class(t *testing.T) {
	html := `<html><body><div id="main" class="active">Content</div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	// 组合 ID 和 class 查询
	nodes := root.QueryNode([]string{"#main", ".active"})
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	}
}

// ==================== HtmNode 方法测试 ====================

func TestHtmNode_Attribute(t *testing.T) {
	html := `<html><body><a href="test-url" title="test-title">link</a></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	nodes := root.QueryNode([]string{"a"})
	if len(nodes) > 0 {
		href := nodes[0].Attribute("href")
		if href != "test-url" {
			t.Errorf("Expected href=test-url, got %s", href)
		}

		title := nodes[0].Attribute("title")
		if title != "test-title" {
			t.Errorf("Expected title=test-title, got %s", title)
		}

		// 不存在的属性
		nonexist := nodes[0].Attribute("nonexist")
		if nonexist != "" {
			t.Error("Non-existent attribute should return empty string")
		}
	}
}

func TestHtmNode_HasClass(t *testing.T) {
	html := `<html><body><div class="foo bar baz">text</div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	nodes := root.QueryNode([]string{"div"})
	if len(nodes) > 0 {
		if !nodes[0].HasClass("foo") {
			t.Error("Should have class foo")
		}
		if !nodes[0].HasClass("bar") {
			t.Error("Should have class bar")
		}
		if !nodes[0].HasClass("baz") {
			t.Error("Should have class baz")
		}
		if nodes[0].HasClass("qux") {
			t.Error("Should not have class qux")
		}
	}
}

func TestHtmNode_HasClass_NotElement(t *testing.T) {
	html := `<html><body>just text</body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	// 文本节点不应该有 class
	if root.HasClass("test") {
		t.Error("Text node should not have class")
	}
}

// ==================== 边界情况测试 ====================

func TestQueryNode_Wildcard(t *testing.T) {
	html := `<html><body><div>1</div><span>2</span><p>3</p></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	// 查询所有元素
	nodes := root.QueryNode([]string{"*"})
	// 通配符查询可能返回所有子节点
	_ = nodes
}

func TestQueryNode_NthChild(t *testing.T) {
	html := `<html><body><ul><li>1</li><li>2</li><li>3</li></ul></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	// 查询特定位置的子元素
	nodes := root.QueryNode([]string{"0"})
	// 数字索引查询
	_ = nodes
}

func TestQueryNode_Sibling(t *testing.T) {
	html := `<html><body><ul><li>1</li><li>2</li><li>3</li></ul></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	// 查询 li + li（兄弟节点）
	nodes := root.QueryNode([]string{"li", "+", "li"})
	// 兄弟节点查询
	_ = nodes
}

// ==================== 性能基准测试 ====================

func BenchmarkParseHtml(b *testing.B) {
	html := `<html><body><div class="container"><ul><li>item1</li><li>item2</li><li>item3</li></ul><p>text</p></div></body></html>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseHtml(strings.NewReader(html))
	}
}

func BenchmarkQueryNode_Simple(b *testing.B) {
	html := `<html><body><div class="container"><p>text</p></div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root.QueryNode([]string{".container"})
	}
}

func BenchmarkQueryNode_Complex(b *testing.B) {
	html := `<html><body><div class="container"><ul><li class="item">1</li><li class="item">2</li><li class="item">3</li><li class="item">4</li><li class="item">5</li></ul></div></body></html>`
	root, _ := ParseHtml(strings.NewReader(html))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root.QueryNode([]string{".container", ".item"})
	}
}