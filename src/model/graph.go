package model

import (
	"container/heap"
	"fmt"
	"math"
	"strings"
	"sync"
)

// Edge represents an edge in the graph.
type Edge struct {
	To     string
	Weight float64
}

// Graph represents a graph with an adjacency list.
type Graph struct {
	adjacencyList map[string][]Edge
	lock          sync.RWMutex // 用于保证并发安全
}

// NewGraph creates a new graph.
func NewGraph() *Graph {
	return &Graph{adjacencyList: make(map[string][]Edge)}
}

func (g *Graph) Flush() {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.adjacencyList = make(map[string][]Edge)
}

// AddEdge adds or updates an edge to the graph.
func (g *Graph) AddEdge(from, to string, weight float64) {
	g.lock.Lock()
	defer g.lock.Unlock()

	// Check if edge already exists and update weight if it does
	for i, edge := range g.adjacencyList[from] {
		if edge.To == to {
			g.adjacencyList[from][i].Weight = weight
			return
		}
	}

	// If edge does not exist, add it
	g.adjacencyList[from] = append(g.adjacencyList[from], Edge{To: to, Weight: weight})
}

func (g *Graph) AddBidirectionalEdge(from, to string, weight float64) {
	g.AddEdge(from, to, weight)
	g.AddEdge(to, from, weight)
}

// RemoveEdge removes an edge from the graph.
func (g *Graph) RemoveEdge(from, to string) {
	g.lock.Lock()
	defer g.lock.Unlock()

	edges := g.adjacencyList[from]
	for i, edge := range edges {
		if edge.To == to {
			g.adjacencyList[from] = append(edges[:i], edges[i+1:]...)
			return
		}
	}
}

func (g *Graph) RemoveBidirectionalEdge(from, to string) {
	g.RemoveEdge(from, to)
	g.RemoveEdge(to, from)
}

// RemoveNode removes a node and all its associated edges from the graph.
func (g *Graph) RemoveNode(node string) {
	g.lock.Lock()
	defer g.lock.Unlock()

	// Remove all edges from this node
	delete(g.adjacencyList, node)

	// Remove all edges to this node
	for from, edges := range g.adjacencyList {
		for i := 0; i < len(edges); {
			if edges[i].To == node {
				edges = append(edges[:i], edges[i+1:]...)
				continue
			}
			i++
		}
		g.adjacencyList[from] = edges
	}
}

func (g *Graph) IterateEdges(from string) []string {
	g.lock.RLock()
	defer g.lock.RUnlock()
	if edges, found := g.adjacencyList[from]; found || len(edges) > 0 {
		dests := make([]string, len(edges))
		for i, edge := range edges {
			dests[i] = edge.To
		}
		return dests
	} else {
		return []string{}
	}
}

func (g *Graph) Print() string {
	g.lock.RLock()
	defer g.lock.RUnlock()
	strList := make([]string, 0, len(g.adjacencyList)*2)
	for from, edges := range g.adjacencyList {
		for _, edge := range edges {
			strList = append(strList, fmt.Sprintf("%s-->|%f|%s\n", from, edge.Weight, edge.To))
		}
	}
	return strings.Join(strList, "\n")
}

func (g *Graph) PrintRoutes(start string, end string) string {
	g.lock.RLock()
	defer g.lock.RUnlock()
	strList := make([]string, 0, len(g.adjacencyList)*2)
	path, latency := g.ShortestPath(start, end)
	path = append(path, end)
	path = append([]string{start}, path...)
	isInPath := make(map[string]map[string]struct{})
	for i := 0; i < len(path)-1; i++ {
		if _, ok := isInPath[path[i]]; !ok {
			isInPath[path[i]] = make(map[string]struct{})
		}
		isInPath[path[i]][path[i+1]] = struct{}{}
	}

	for from, edges := range g.adjacencyList {
		for _, edge := range edges {
			if _, ok := isInPath[from][edge.To]; ok {
				strList = append(strList, fmt.Sprintf("%s == \"%f\" ==> %s\n", from, edge.Weight, edge.To))
			} else {
				strList = append(strList, fmt.Sprintf("%s-. \"%f\" .-> %s\n", from, edge.Weight, edge.To))
			}
		}
	}
	strList = append(strList, fmt.Sprintf("lat(Latency: %f)\n", latency))
	return strings.Join(strList, "\n")
}

// Define a structure for items in the priority queue
type Item struct {
	node     string
	distance float64
	index    int // The index is needed by update and is maintained by the heap.Interface methods.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].distance < pq[j].distance // The priority is the distance; lower distance means higher priority.
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (g *Graph) ShortestPath(start, target string) ([]string, float64) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	dist := make(map[string]float64)
	prev := make(map[string]string)

	for node, edges := range g.adjacencyList {
		dist[node] = math.Inf(1)
		for _, edge := range edges {
			dist[edge.To] = math.Inf(1)
		}
	}
	dist[start] = 0

	pq := make(PriorityQueue, 0)
	heap.Push(&pq, &Item{node: start, distance: 0})

	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		currentNode := item.node

		if currentNode == target {
			break
		}

		for _, edge := range g.adjacencyList[currentNode] {
			alt := dist[currentNode] + edge.Weight
			if alt < dist[edge.To] {
				dist[edge.To] = alt
				prev[edge.To] = currentNode
				heap.Push(&pq, &Item{node: edge.To, distance: alt})
			}
		}
	}

	path := make([]string, 0)
	u, ok := prev[target]
	if !ok {
		return nil, math.Inf(1) // Target node is not reachable from start
	}
	for u != start {
		path = append([]string{u}, path...)
		u = prev[u]
	}

	return path, dist[target]
}
