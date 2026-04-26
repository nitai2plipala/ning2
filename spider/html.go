package spider

import (
	"golang.org/x/net/html"
	"io"
	"log"
	"strings"
)

/*
	ErrorNode         0
	TextNode          1
	Document          2
	Element           3      默认值
	Comment           4
	Doctype           5
	ScopeMarker       6

	["0-6", "css: tagname #id .class [attribute]", "+ > *" ]

	: Child Only First Last Nth Not Range

	Element Tagname a-zA-Z  #id  .class  [attribute]

	后代（默认）  * 所有   + 之后   > 子代   , 任满足其一

	[ attribute ]  name value  =  ~  |  ^  $  *

*/



// copyNode 递归复制 html.Node 为 HtmNode，parent 参数用于设置子节点的父节点
func copyNodeWithParent(n *html.Node, parent *HtmNode) *HtmNode {
	if n == nil {
		return nil
	}

	// 递归复制子节点
	var firstChild, lastChild *HtmNode
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		child := copyNodeWithParent(c, nil) // 先创建子节点，父节点稍后设置
		if firstChild == nil {
			firstChild = child
			lastChild = child
		} else {
			lastChild.NextSibling = child
			child.PrevSibling = lastChild
			lastChild = child
		}
	}

	node := &HtmNode{
		Parent:      parent,
		FirstChild:  firstChild,
		LastChild:   lastChild,
		PrevSibling: nil,
		NextSibling: nil,
		Type:        n.Type,
		DataAtom:    n.DataAtom,
		Data:        n.Data,
		Namespace:   n.Namespace,
		Attr:        copyAttributes(n.Attr),
	}

	// 设置子节点的父节点
	for c := firstChild; c != nil; c = c.NextSibling {
		c.Parent = node
	}

	return node
}

// copyNode 递归复制 html.Node 为 HtmNode
func copyNode(n *html.Node) *HtmNode {
	return copyNodeWithParent(n, nil)
}

// copyAttributes 复制属性切片
func copyAttributes(attrs []html.Attribute) []html.Attribute {
	if attrs == nil {
		return nil
	}
	result := make([]html.Attribute, len(attrs))
	copy(result, attrs)
	return result
}

// exactMatch 检查属性值是否精确包含目标值（用于 ID/Class 精确匹配）
func exactMatch(attrVal, target string) bool {
	for _, v := range strings.Split(attrVal, " ") {
		if v == target {
			return true
		}
	}
	return false
}

func ParseHtml(reader io.Reader) (*HtmNode, error) {
	node, err := html.Parse(reader)
	if err != nil {
		return nil, err
	}

	// 复制完整节点树
	return copyNode(node), nil
}


func (t *HtmNode) QueryNode(pattern []string) []*HtmNode {
	if len(pattern) == 0 {
		return nil
	}

	buffer, intake := t.NodeFind(pattern)

	for intake > 0 && intake < len(pattern) {
		pattern = pattern[intake:]

		var htmNode []*HtmNode

		for _, n := range buffer {
			node, newIntake := n.NodeFind(pattern)
			htmNode = append(htmNode, node...)
			// 更新 intake（取最大值）
			if newIntake > intake {
				intake = newIntake
			}
		}

		buffer = htmNode
	}

	return buffer
}

func (t *HtmNode) NodeFind (pattern []string) ([]*HtmNode, int) {

	switch pattern[0] {

	case "*":

		return t.htmBuffer(nil, []string{"*"}), 1

	case "0", "1", "2", "3", "4":

		return t.htmBuffer(nil, []string{pattern[0]}), 1

	case "+", ">":

		return t.htmBuffer(nil, pattern[:2]), 2

	case ",":

		log.Println(", catch error")

		return nil, 0

	default :

	    return t.htmBuffer(nil, []string{pattern[0]}), 1
	}

}


func (t *HtmNode) htmBuffer (buffer []*HtmNode, pattern []string) []*HtmNode {

	//fmt.Println(buffer, t)

	//fmt.Println("Len:", len(pattern), pattern)

	switch pattern[0] {

	case ">" :  //    > 子代

	    //fmt.Println(pattern, t)

		for c := t.FirstChild; c != nil; c = c.NextSibling {

			var count int = 0

			for _, value := range pattern[1:] {

				switch value {

				case "0", "1", "2", "3", "4":

				default:

					switch value[:1] {

					case "#":    //   ID

						if c.Type == html.ElementNode {

							for _, attr := range c.Attr {

								if attr.Key == "id" {

									if exactMatch(attr.Val, value[1:]) {

										count++
									}

								}
							}

						}

					case ".":     //   CLASS

						if c.Type == html.ElementNode {

							for _, attr := range c.Attr {

								if attr.Key == "class" {

									if exactMatch(attr.Val, value[1:]) {

										count++
									}

								}
							}

						}

					case "[" :     //   Attribute

						if c.Type == html.ElementNode && len(value) > 2 {
							attrStr := value[1 : len(value)-1]
							parts := strings.SplitN(attrStr, "=", 2)
							attrName := parts[0]
							var attrVal string
							if len(parts) > 1 {
								attrVal = parts[1]
								attrVal = strings.Trim(attrVal, "\"'")
							}

							for _, attr := range c.Attr {
								if attr.Key == attrName {
									if len(parts) == 1 || attr.Val == attrVal {
										count++
										break
									}
								}
							}
						}

					default :      //   TAG

						if c.Type == html.ElementNode {

							if c.Data == value {

								count++
							}
						}

					}

				}
			}

			if count == len(pattern)-1 {
				// 复制节点（需要将 *html.Node 转换为 *HtmNode）
				node := &HtmNode{
					Parent:      t,
					FirstChild:  nil, // 子节点在查询结果中不需要
					LastChild:   nil,
					PrevSibling: nil,
					NextSibling: nil,
					Type:        c.Type,
					DataAtom:    c.DataAtom,
					Data:        c.Data,
					Namespace:   c.Namespace,
					Attr:        copyAttributes(c.Attr),
				}

				buffer = append(buffer, node)
			}

		}

		//fmt.Println(buffer)

		return buffer

	case "+" :  //    + 之后（兄弟节点）
		// 查找当前节点的下一个兄弟节点
		sibling := t.NextSibling
		for sibling != nil {
			var count int = 0

			for _, value := range pattern[1:] {

				switch value {

				case "0", "1", "2", "3", "4":

				default:

					switch value[:1] {

					case "#":    //   ID

						if sibling.Type == html.ElementNode {

							for _, attr := range sibling.Attr {

								if attr.Key == "id" {

									if exactMatch(attr.Val, value[1:]) {

										count++
									}

								}
							}

						}

					case ".":     //   CLASS

						if sibling.Type == html.ElementNode {

							for _, attr := range sibling.Attr {

								if attr.Key == "class" {

									if exactMatch(attr.Val, value[1:]) {

										count++
									}

								}
							}

						}

					case "[" :     //   Attribute

						if sibling.Type == html.ElementNode && len(value) > 2 {
							attrStr := value[1 : len(value)-1]
							parts := strings.SplitN(attrStr, "=", 2)
							attrName := parts[0]
							var attrVal string
							if len(parts) > 1 {
								attrVal = parts[1]
								attrVal = strings.Trim(attrVal, "\"'")
							}

							for _, attr := range sibling.Attr {
								if attr.Key == attrName {
									if len(parts) == 1 || attr.Val == attrVal {
										count++
										break
									}
								}
							}
						}

					default :      //   TAG

						if sibling.Type == html.ElementNode {

							if sibling.Data == value {

								count++
							}
						}

					}

				}
			}

			if count == len(pattern)-1 {
				// 复制节点
				node := &HtmNode{
					Parent:      t.Parent,
					FirstChild:  nil,
					LastChild:   nil,
					PrevSibling: nil,
					NextSibling: nil,
					Type:        sibling.Type,
					DataAtom:    sibling.DataAtom,
					Data:        sibling.Data,
					Namespace:   sibling.Namespace,
					Attr:        copyAttributes(sibling.Attr),
				}

				buffer = append(buffer, node)
			}

			// 只看下一个兄弟节点
			break
		}

		return buffer

	case "*" :  //    * 所有

	case "," :  //    , 任满足其一

	case "0", "1", "2", "3", "4":


	default :   //      后代（默认）

	    var count int = 0

		for _, value := range pattern {

			switch value {

			case "0", "1", "2", "3", "4":

			default:

				switch value[:1] {

				case "#":    //   ID

				    if t.Type == html.ElementNode {

				    	for _, attr := range t.Attr {

				    		if attr.Key == "id" {

				    			if exactMatch(attr.Val, value[1:]) {

									count++
								}
							}
						}

					}

				case ".":     //   CLASS

					if t.Type == html.ElementNode {

						for _, attr := range t.Attr {

							if attr.Key == "class" {

								if exactMatch(attr.Val, value[1:]) {

									count++
								}

							}
						}

					}

				case "[" :     //   Attribute

					if t.Type == html.ElementNode && len(value) > 2 {
						attrStr := value[1 : len(value)-1]
						parts := strings.SplitN(attrStr, "=", 2)
						attrName := parts[0]
						var attrVal string
						if len(parts) > 1 {
							attrVal = parts[1]
							attrVal = strings.Trim(attrVal, "\"'")
						}

						for _, attr := range t.Attr {
							if attr.Key == attrName {
								if len(parts) == 1 || attr.Val == attrVal {
									count++
									break
								}
							}
						}
					}

				default :      //   TAG

					if t.Type == html.ElementNode {

						if t.Data == value {

							count++
						}
					}

				}
				
			}
		}

		if count == len(pattern) {

			//fmt.Println(t)

			buffer = append(buffer, t)
		}
	}


	for c := t.FirstChild; c != nil; c = c.NextSibling {
		// 复制节点（需要将 *html.Node 转换为 *HtmNode）
		// 注意：这里需要保留子节点引用，否则递归无法遍历
		node := &HtmNode{
			Parent:      t,
			FirstChild:  c.FirstChild,
			LastChild:   c.LastChild,
			PrevSibling: nil,
			NextSibling: nil,
			Type:        c.Type,
			DataAtom:    c.DataAtom,
			Data:        c.Data,
			Namespace:   c.Namespace,
			Attr:        copyAttributes(c.Attr),
		}

		buffer = node.htmBuffer(buffer, pattern)
	}

	//fmt.Println("buffer:", buffer)

	return buffer
}


func (t *HtmNode) Attribute (name string) string {

	if t.Type == html.ElementNode {

		for _, attr := range t.Attr {

			if attr.Key == name {

				return attr.Val
			}
		}

	}

	return ""
}

func (t *HtmNode) HasClass(name string) bool {
	if t.Type == html.ElementNode {
		for _, attr := range t.Attr {
			if attr.Key == "class" {
				if exactMatch(attr.Val, name) {
					return true
				}
			}
		}
	}
	return false
}


