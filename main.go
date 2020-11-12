package main

import (
	"container/heap"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type Edge struct {
	in     int
	out    int
	weight float64
}

type Node struct {
	edges []*Edge
	value NodeValue
}

func (n *Node) addEdge(from_idx int, to_idx int, weight float64) {
	n.edges = append(n.edges, &Edge{from_idx, to_idx, weight})
}

type NodeValue struct {
	x float64
	y float64
}

type Graph struct {
	nodes []*Node
}

func (g *Graph) addNode(node *Node) int {
	g.nodes = append(g.nodes, node)
	return len(g.nodes) - 1
}

func (g *Graph) addEdgeBidirectional(from_idx int, to_idx int, weight float64) {
	g.addEdge(from_idx, to_idx, weight)
	g.addEdge(to_idx, from_idx, weight)
}

func (g *Graph) addEdge(from_idx int, to_idx int, weight float64) {
	node := g.nodes[from_idx]
	node.addEdge(from_idx, to_idx, weight)
}

func (g *Graph) print() {
	fmt.Println("Nodes: ", len(g.nodes))
	for i, n := range g.nodes {
		fmt.Println(i, ": ", n.value)
		for _, e := range n.edges {
			fmt.Println(e.in, " -> ", e.out, " ", e.weight)
		}
	}
}

func euclidean(a NodeValue, b NodeValue) float64 {
	return math.Sqrt(math.Pow(a.x-b.x, 2) + math.Pow(a.y-b.y, 2))
}

type AStarPQElement struct {
	curr_index   int
	prev_index   int
	cost_to_come float64
	cost_to_go   float64
	Index        int
}

type AStarVisitedElement struct {
	prev_idx     int
	cost_to_come float64
}

type AStarResult struct {
	path          []int
	path_len      float64
	visited_count int
}

type PriorityQueue []*AStarPQElement

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].cost_to_come+pq[i].cost_to_go < pq[j].cost_to_come+pq[j].cost_to_go
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*AStarPQElement)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func reverse(a []int) []int {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
	return a
}

func (g *Graph) extractAStarSolution(start_idx int, goal_idx int, visited map[int]AStarVisitedElement, num_visited int) AStarResult {
	path := []int{goal_idx}
	prev, _ := visited[goal_idx]
	pathlen := prev.cost_to_come
	// this is poorly written lol
	for prev.prev_idx != start_idx {
		curr := prev
		path = append(path, curr.prev_idx)
		prev, _ = visited[curr.prev_idx]
	}
	path = append(path, start_idx)

	return AStarResult{reverse(path), pathlen, num_visited}
}

func (g *Graph) AStarSearch(start_idx int, goal_idx int, h func(NodeValue, NodeValue) float64) AStarResult {

	start := g.nodes[start_idx].value
	goal := g.nodes[goal_idx].value

	// Create priority queue
	pq := PriorityQueue{}
	heap.Init(&pq)

	// Create visited map
	visited := make(map[int]AStarVisitedElement)

	numVisited := 0

	heap.Push(&pq, &AStarPQElement{start_idx, -1, 0.0, h(start, goal), 0})

	for pq.Len() > 0 {
		best := heap.Pop(&pq).(*AStarPQElement)
		numVisited++

		seen, found := visited[best.curr_index]
		better := false
		if found && best.cost_to_come < seen.cost_to_come {
			better = true
		}
		if !found || better {
			visited[best.curr_index] = AStarVisitedElement{best.prev_index, best.cost_to_come}

			if best.curr_index == goal_idx {
				break
			}

			parent := g.nodes[best.curr_index]
			for _, edge := range parent.edges {
				childCostToCome := best.cost_to_come + edge.weight
				child, childFound := visited[edge.out]
				childBetter := false
				if childFound && childCostToCome < child.cost_to_come {
					childBetter = true
				}

				if !childFound || childBetter {
					heap.Push(&pq, &AStarPQElement{edge.out, best.curr_index, childCostToCome, h(g.nodes[edge.out].value, goal), 0})
				}
			}
		}
	}
	return g.extractAStarSolution(start_idx, goal_idx, visited, numVisited)
}

func main() {
	g := Graph{}

	node_file, err := os.Open("./data/cal.cnode")
	if err != nil {
		log.Fatalln("Couldn't open the node file", err)
	}
	r := csv.NewReader(node_file)
	for {
		// seems like a bad pattern
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// also seems bad lol
		record = strings.Split(record[0], " ")

		long, _ := strconv.ParseFloat(record[1], 64)
		lat, _ := strconv.ParseFloat(record[2], 64)

		g.addNode(&Node{[]*Edge{}, NodeValue{lat, long}})
	}

	edge_file, err := os.Open("./data/cal.cedge")
	if err != nil {
		log.Fatalln("Couldn't open the edge file", err)
	}
	r = csv.NewReader(edge_file)
	for {
		// seems like a bad pattern
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// also seems bad lol
		record = strings.Split(record[0], " ")
		idx_one, _ := strconv.ParseInt(record[1], 10, 32)
		idx_two, _ := strconv.ParseInt(record[2], 10, 32)
		weight, _ := strconv.ParseFloat(record[3], 64)
		g.addEdgeBidirectional(int(idx_one), int(idx_two), weight)
	}

	start := time.Now()
	res := g.AStarSearch(7261, 20286, euclidean)
	fmt.Println("Finished search in : ", time.Now().Sub(start))
	fmt.Println("Nodes visited: ", res.visited_count)
	fmt.Println("Path len: ", res.path_len)
	fmt.Println("Path: ", res.path)

}
