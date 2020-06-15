package mux

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

const maxParamCount = ^uint8(0)

func countParams(path string) uint8 {
	var n uint
	for i := 0; i < len(path); i++ {
		if path[i] != ':' && path[i] != '*' {
			continue
		}
		n++
	}
	if n >= uint(maxParamCount) {
		return maxParamCount
	}

	return uint8(n)
}

type nodeType uint8

const (
	static nodeType = iota // default
	root
	param
	catchAll
)

type node struct {
	path      string
	wildChild bool
	nType     nodeType
	maxParams uint8
	priority  uint32
	indices   string
	children  []*node
	handle    HandlerFunc
}

// increments priority of the given child and reorders if necessary
func (n *node) incrementChildPriority(pos int) int {
	n.children[pos].priority++
	priority := n.children[pos].priority

	// adjust position (move to front)
	newPos := pos
	for newPos > 0 && n.children[newPos-1].priority < priority {
		// swap node positions
		n.children[newPos-1], n.children[newPos] = n.children[newPos], n.children[newPos-1]
		newPos--
	}

	// build new index char string
	if newPos != pos {
		n.indices = n.indices[:newPos] + // unchanged prefix, might be empty
			n.indices[pos:pos+1] + // the index char we move
			n.indices[newPos:pos] + n.indices[pos+1:] // rest without char at 'pos'
	}

	return newPos
}

// addRoute adds a node with the given handle to the path.
// Not concurrency-safe!
func (n *node) addRoute(path string, handle HandlerFunc) {
	fullPath := path
	n.priority++
	numParams := countParams(path)

	// non-empty tree
	if len(n.path) > 0 || len(n.children) > 0 {
		now := n
	walk:
		for {
			// Update maxParams of the current node
			if numParams > now.maxParams {
				now.maxParams = numParams
			}

			// Find the longest common prefix.
			// This also implies that the common prefix contains no ':' or '*'
			// since the existing key can't contain those chars.
			i := 0
			max := min(len(path), len(now.path))
			for i < max && path[i] == now.path[i] {
				i++
			}

			// Split edge
			if i < len(now.path) {
				child := node{
					path:      now.path[i:],
					wildChild: now.wildChild,
					nType:     static,
					indices:   now.indices,
					children:  now.children,
					handle:    now.handle,
					priority:  now.priority - 1,
				}

				// Update maxParams (max of all children)
				for i := range child.children {
					if child.children[i].maxParams > child.maxParams {
						child.maxParams = child.children[i].maxParams
					}
				}

				now.children = []*node{&child}
				// []byte for proper unicode char conversion, see #65
				now.indices = string([]byte{now.path[i]})
				now.path = path[:i]
				now.handle = nil
				now.wildChild = false
			}

			// Make new node a child of this node
			if i < len(path) {
				path = path[i:]

				if now.wildChild {
					now = now.children[0]
					now.priority++

					// Update maxParams of the child node
					if numParams > now.maxParams {
						now.maxParams = numParams
					}
					numParams--

					// Check if the wildcard matches
					if len(path) >= len(now.path) && now.path == path[:len(now.path)] &&
						// Adding a child to a catchAll is not possible
						now.nType != catchAll &&
						// Check for longer wildcard, e.g. :name and :namec
						(len(now.path) >= len(path) || path[len(now.path)] == '/') {
						continue walk
					} else {
						// Wildcard conflict
						var pathSeg string
						if now.nType == catchAll {
							pathSeg = path
						} else {
							pathSeg = strings.SplitN(path, "/", 2)[0]
						}
						prefix := fullPath[:strings.Index(fullPath, pathSeg)] + now.path
						panic("'" + pathSeg +
							"' in new path '" + fullPath +
							"' conflicts with existing wildcard '" + now.path +
							"' in existing prefix '" + prefix +
							"'")
					}
				}

				c := path[0]

				// slash after param
				if now.nType == param && c == '/' && len(now.children) == 1 {
					now = now.children[0]
					now.priority++
					continue walk
				}

				// Check if a child with the next path byte exists
				for i := 0; i < len(now.indices); i++ {
					if c == now.indices[i] {
						i = now.incrementChildPriority(i)
						now = now.children[i]
						continue walk
					}
				}

				// Otherwise insert it
				if c != ':' && c != '*' {
					// []byte for proper unicode char conversion, see #65
					now.indices += string([]byte{c})
					child := &node{
						maxParams: numParams,
					}
					now.children = append(now.children, child)
					now.incrementChildPriority(len(now.indices) - 1)
					now = child
				}
				now.insertChild(numParams, path, fullPath, handle)
				return

			} else if i == len(path) { // Make node a (in-path) leaf
				if now.handle != nil {
					panic("a handle is already registered for path '" + fullPath + "'")
				}
				now.handle = handle
			}
			return
		}
	} else { // Empty tree
		n.insertChild(numParams, path, fullPath, handle)
		n.nType = root
	}
}

func (n *node) insertChild(numParams uint8, path, fullPath string, handle HandlerFunc) {
	var offset int // already handled bytes of the path

	now := n

	// find prefix until first wildcard (beginning with ':'' or '*'')
	for i, max := 0, len(path); numParams > 0; i++ {
		c := path[i]
		if c != ':' && c != '*' {
			continue
		}

		// find wildcard end (either '/' or path end)
		end := i + 1
		for end < max && path[end] != '/' {
			switch path[end] {
			// the wildcard name must not contain ':' and '*'
			case ':', '*':
				panic("only one wildcard per path segment is allowed, has: '" +
					path[i:] + "' in path '" + fullPath + "'")
			default:
				end++
			}
		}

		// check if this Node existing children which would be
		// unreachable if we insert the wildcard here
		if len(now.children) > 0 {
			panic("wildcard route '" + path[i:end] +
				"' conflicts with existing children in path '" + fullPath + "'")
		}

		// check if the wildcard has a name
		if end-i < 2 {
			panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
		}

		if c == ':' { // param
			// split path at the beginning of the wildcard
			if i > 0 {
				now.path = path[offset:i]
				offset = i
			}

			child := &node{
				nType:     param,
				maxParams: numParams,
			}
			now.children = []*node{child}
			now.wildChild = true
			now = child
			now.priority++
			numParams--

			// if the path doesn't end with the wildcard, then there
			// will be another non-wildcard subpath starting with '/'
			if end < max {
				now.path = path[offset:end]
				offset = end

				child := &node{
					maxParams: numParams,
					priority:  1,
				}
				now.children = []*node{child}
				now = child
			}

		} else { // catchAll
			if end != max || numParams > 1 {
				panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
			}

			if len(now.path) > 0 && now.path[len(now.path)-1] == '/' {
				panic("catch-all conflicts with existing handle for the path segment root in path '" + fullPath + "'")
			}

			// currently fixed width 1 for '/'
			i--
			if path[i] != '/' {
				panic("no / before catch-all in path '" + fullPath + "'")
			}

			now.path = path[offset:i]

			// first node: catchAll node with empty path
			child := &node{
				wildChild: true,
				nType:     catchAll,
				maxParams: 1,
			}
			// update maxParams of the parent node
			if now.maxParams < 1 {
				now.maxParams = 1
			}
			now.children = []*node{child}
			now.indices = string(path[i])
			now = child
			now.priority++

			// second node: node holding the variable
			child = &node{
				path:      path[i:],
				nType:     catchAll,
				maxParams: 1,
				handle:    handle,
				priority:  1,
			}
			now.children = []*node{child}

			return
		}
	}

	// insert remaining path part and handle to the leaf
	now.path = path[offset:]
	now.handle = handle
}

// Returns the handle registered with the given path (key). The values of
// wildcards are saved to a map.
// If no handle can be found, a TSR (trailing slash redirect) recommendation is
// made if a handle exists with an extra (without the) trailing slash for the
// given path.
func (n *node) getValue(path string) (handle HandlerFunc, p map[string]string, tsr bool) {
	p = make(map[string]string)
	now := n
walk: // outer loop for walking the tree
	for {
		if len(path) > len(now.path) {
			if path[:len(now.path)] == now.path {
				path = path[len(now.path):]
				// If this node does not have a wildcard (param or catchAll)
				// child,  we can just look up the next child node and continue
				// to walk down the tree
				if !now.wildChild {
					c := path[0]
					for i := 0; i < len(now.indices); i++ {
						if c == now.indices[i] {
							now = now.children[i]
							continue walk
						}
					}

					// Nothing found.
					// We can recommend to redirect to the same URL without a
					// trailing slash if a leaf exists for that path.
					tsr = path == "/" && now.handle != nil
					return
				}

				// handle wildcard child
				now = now.children[0]
				switch now.nType {
				case param:
					// find param end (either '/' or path end)
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}
					p[now.path[1:]] = path[:end]

					// we need to go deeper!
					if end < len(path) {
						if len(now.children) > 0 {
							path = path[end:]
							now = now.children[0]
							continue walk
						}

						// ... but we can't
						tsr = len(path) == end+1
						return
					}

					if handle = now.handle; handle != nil {
						return
					} else if len(now.children) == 1 {
						// No handle found. Check if a handle for this path + a
						// trailing slash exists for TSR recommendation
						now = now.children[0]
						tsr = now.path == "/" && now.handle != nil
					}

					return
				case catchAll:
					p[now.path[2:]] = path
					handle = now.handle
					return
				default:
					panic("invalid node type")
				}
			}
		} else if path == now.path {
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if handle = now.handle; handle != nil {
				return
			}

			if path == "/" && now.wildChild && now.nType != root {
				tsr = true
				return
			}

			// No handle found. Check if a handle for this path + a
			// trailing slash exists for trailing slash recommendation
			for i := 0; i < len(now.indices); i++ {
				if now.indices[i] == '/' {
					now = now.children[i]
					tsr = (len(now.path) == 1 && now.handle != nil) ||
						(now.nType == catchAll && now.children[0].handle != nil)
					return
				}
			}

			return
		}

		// Nothing found. We can recommend to redirect to the same URL with an
		// extra trailing slash if a leaf exists for that path
		tsr = (path == "/") ||
			(len(now.path) == len(path)+1 && now.path[len(path)] == '/' &&
				path == now.path[:len(now.path)-1] && now.handle != nil)
		return
	}
}

// Makes a case-insensitive lookup of the given path and tries to find a handler.
// It can optionally also fix trailing slashes.
// It returns the case-corrected path and a bool indicating whether the lookup
// was successful.
func (n *node) findCaseInsensitivePath(path string, fixTrailingSlash bool) (ciPath []byte, found bool) {
	return n.findCaseInsensitivePathRec(
		path,
		make([]byte, 0, len(path)+1), // preallocate enough memory for new path
		[4]byte{},                    // empty rune buffer
		fixTrailingSlash,
	)
}

// shift bytes in array by n bytes left
func shiftNRuneBytes(rb [4]byte, n int) [4]byte {
	switch n {
	case 0:
		return rb
	case 1:
		return [4]byte{rb[1], rb[2], rb[3], 0}
	case 2:
		return [4]byte{rb[2], rb[3]}
	case 3:
		return [4]byte{rb[3]}
	default:
		return [4]byte{}
	}
}

// recursive case-insensitive lookup function used by n.findCaseInsensitivePath
func (n *node) findCaseInsensitivePathRec(path string, ciPath []byte, rb [4]byte, fixTrailingSlash bool) ([]byte, bool) {
	npLen := len(n.path)

	now := n

walk: // outer loop for walking the tree
	for len(path) >= npLen && (npLen == 0 || strings.EqualFold(path[1:npLen], now.path[1:])) {
		// add common prefix to result

		oldPath := path
		path = path[npLen:]
		ciPath = append(ciPath, now.path...)

		if len(path) > 0 {
			// If this node does not have a wildcard (param or catchAll) child,
			// we can just look up the next child node and continue to walk down
			// the tree
			if !now.wildChild {
				// skip rune bytes already processed
				rb = shiftNRuneBytes(rb, npLen)

				if rb[0] != 0 {
					// old rune not finished
					for i := 0; i < len(now.indices); i++ {
						if now.indices[i] == rb[0] {
							// continue with child node
							now = now.children[i]
							npLen = len(now.path)
							continue walk
						}
					}
				} else {
					// process a new rune
					var rv rune

					// find rune start
					// runes are up to 4 byte long,
					// -4 would definitely be another rune
					var off int
					for max := min(npLen, 3); off < max; off++ {
						if i := npLen - off; utf8.RuneStart(oldPath[i]) {
							// read rune from cached path
							rv, _ = utf8.DecodeRuneInString(oldPath[i:])
							break
						}
					}

					// calculate lowercase bytes of current rune
					lo := unicode.ToLower(rv)
					utf8.EncodeRune(rb[:], lo)

					// skip already processed bytes
					rb = shiftNRuneBytes(rb, off)

					for i := 0; i < len(now.indices); i++ {
						// lowercase matches
						if now.indices[i] == rb[0] {
							// must use a recursive approach since both the
							// uppercase byte and the lowercase byte might exist
							// as an index
							if out, found := now.children[i].findCaseInsensitivePathRec(
								path, ciPath, rb, fixTrailingSlash,
							); found {
								return out, true
							}
							break
						}
					}

					// if we found no match, the same for the uppercase rune,
					// if it differs
					if up := unicode.ToUpper(rv); up != lo {
						utf8.EncodeRune(rb[:], up)
						rb = shiftNRuneBytes(rb, off)

						for i, c := 0, rb[0]; i < len(now.indices); i++ {
							// uppercase matches
							if now.indices[i] == c {
								// continue with child node
								now = now.children[i]
								npLen = len(now.path)
								continue walk
							}
						}
					}
				}

				// Nothing found. We can recommend to redirect to the same URL
				// without a trailing slash if a leaf exists for that path
				return ciPath, fixTrailingSlash && path == "/" && now.handle != nil
			}

			now = now.children[0]
			switch now.nType {
			case param:
				// find param end (either '/' or path end)
				k := 0
				for k < len(path) && path[k] != '/' {
					k++
				}

				// add param value to case insensitive path
				ciPath = append(ciPath, path[:k]...)

				// we need to go deeper!
				if k < len(path) {
					if len(now.children) > 0 {
						// continue with child node
						now = now.children[0]
						npLen = len(now.path)
						path = path[k:]
						continue
					}

					// ... but we can't
					if fixTrailingSlash && len(path) == k+1 {
						return ciPath, true
					}
					return ciPath, false
				}

				if now.handle != nil {
					return ciPath, true
				} else if fixTrailingSlash && len(now.children) == 1 {
					// No handle found. Check if a handle for this path + a
					// trailing slash exists
					now = now.children[0]
					if now.path == "/" && now.handle != nil {
						return append(ciPath, '/'), true
					}
				}
				return ciPath, false

			case catchAll:
				return append(ciPath, path...), true

			default:
				panic("invalid node type")
			}
		} else {
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if now.handle != nil {
				return ciPath, true
			}

			// No handle found.
			// Try to fix the path by adding a trailing slash
			if fixTrailingSlash {
				for i := 0; i < len(now.indices); i++ {
					if now.indices[i] == '/' {
						now = now.children[i]
						if (len(now.path) == 1 && now.handle != nil) ||
							(now.nType == catchAll && now.children[0].handle != nil) {
							return append(ciPath, '/'), true
						}
						return ciPath, false
					}
				}
			}
			return ciPath, false
		}
	}

	// Nothing found.
	// Try to fix the path by adding / removing a trailing slash
	if fixTrailingSlash {
		if path == "/" {
			return ciPath, true
		}
		if len(path)+1 == npLen && now.path[len(path)] == '/' &&
			strings.EqualFold(path[1:], now.path[1:len(path)]) && now.handle != nil {
			return append(ciPath, now.path...), true
		}
	}
	return ciPath, false
}
