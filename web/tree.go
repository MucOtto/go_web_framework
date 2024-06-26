package web

import "strings"

// uri的前缀树

type treeNode struct {
	val        string
	children   []*treeNode
	routerName string
	isEnd      bool
}

func (t *treeNode) Put(path string) {
	root := t
	strs := strings.Split(path, "/")
	for index, str := range strs {
		// 刚开始等于空的情况
		if index == 0 {
			continue
		}
		children := t.children
		isMatch := false
		for _, node := range children {
			if node.val == str {
				isMatch = true
				t = node
				break
			}
		}
		if !isMatch {
			isEnd := false
			if index == len(strs)-1 {
				isEnd = true
			}
			node := &treeNode{
				val:      str,
				children: make([]*treeNode, 0),
				isEnd:    isEnd,
			}
			t.children = append(children, node)
			t = node
		}
	}
	t = root
}

func (t *treeNode) Get(path string) *treeNode {
	strs := strings.Split(path, "/")
	routerName := ""
	for index, val := range strs {
		if index == 0 {
			continue
		}
		children := t.children
		isMatch := false
		for _, node := range children {
			if node.val == val ||
				node.val == "*" ||
				strings.Contains(node.val, ":") {
				isMatch = true
				routerName += "/" + node.val
				node.routerName = routerName
				t = node
				if index == len(strs)-1 {
					return node
				}
				break
			}
		}
		if !isMatch {
			for _, node := range children {
				// /user/**
				// /user/get/userInfo
				// /user/aa/bb
				if node.val == "**" {
					routerName += "/" + node.val
					node.routerName = routerName
					return node
				}
			}
		}
	}
	return nil
}
