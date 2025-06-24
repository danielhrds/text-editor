package main

import (
	"fmt"
	"iter"
)

type Node[T any] struct {
	Value    T
	Next     *Node[T]
	Previous *Node[T]
}

type LinkedList[T any] struct {
	Length int
	First  *Node[T]
	Last   *Node[T]
}

func NewLinkedList[T any](value T) LinkedList[T] {
	newNode := &Node[T]{Value: value}
	return LinkedList[T]{
		First:  newNode,
		Last:   newNode,
		Length: 1,
	}
}

func (ll *LinkedList[T]) GetNodeAt(index int) (*Node[T], error) {
	if index > int(ll.Length) {
		return nil, fmt.Errorf("GetNodeAt: error trying to get node. index > Length")
	}
	nodeAtIndex := ll.First
	var i int
	for nodeAtIndex != nil {
		if i == index {
			break
		}
		nodeAtIndex = nodeAtIndex.Next
		i++
	}
	return nodeAtIndex, nil
}

func (ll *LinkedList[T]) Size() int {
	return ll.Length
}

func (ll *LinkedList[T]) Append(value T) {
	newNode := &Node[T]{Value: value}
	newNode.Previous = ll.Last // newNode: next nil previous have a value
	ll.Last.Next = newNode     // current last: next have a value and next too
	ll.Last = newNode          // current last: doesn't have a next since newNode took it's place
	ll.Length++
}

func (ll *LinkedList[T]) InsertAt(value T, index int) error {
	if index > int(ll.Length) {
		return fmt.Errorf("InsertAt: error trying to insert. index > Length")
	}
	newNode := &Node[T]{Value: value}
	if index == 0 {
		previousFirst := ll.First
		ll.First = newNode
		newNode.Next = previousFirst
		previousFirst.Previous = ll.First
	}
	if index > 0 && index < int(ll.Length) {
		// TODO: find if searching forward or backwards is more rapid
		nodeAtIndex, err := ll.GetNodeAt(index)
		if err != nil {
			return err
		}
		oldPrevious := nodeAtIndex.Previous
		oldPrevious.Next = newNode 					// oldPrevious -> newNode
		newNode.Next = nodeAtIndex					// oldPrevious -> newNode -> nodeAtIndex
		newNode.Previous = oldPrevious			// oldPrevious <- newNode
		nodeAtIndex.Previous = newNode			// oldPrevious <- newNode <- nodeAtIndex
	}
	if index == int(ll.Length) {
		ll.Append(value)
		return nil // since Append already increase Length it's safe to return
	}
	ll.Length++
	return nil
}

func (ll *LinkedList[T]) Pop() {
	ll.Last = ll.Last.Previous
	ll.Last.Next = nil
	ll.Length--
}

func (ll *LinkedList[T]) DeleteAt(index int) error {
	// TODO: find if searching forward or backwards is more rapid
	if index > int(ll.Length) {
		return fmt.Errorf("DeleteAt: error trying to delete. index > Length")
	}

	if index == 0 {
		ll.First = ll.First.Next
		ll.First.Previous = nil
	}
	if index > 0 && index < int(ll.Length)-1 {
		nodeAtIndex, err := ll.GetNodeAt(index)
		if err != nil {
			return err
		}
		previous := nodeAtIndex.Previous
		next := nodeAtIndex.Next
		previous.Next = next
		next.Previous = previous
	}
	if index == int(ll.Length)-1 {
		ll.Pop()
		return nil // since Pop already decreases Length it's safe to return
	}
	ll.Length--
	return nil
}

func (ll *LinkedList[T]) GetAt(index int) (T, error) {
	if index > int(ll.Length) {
		return *new(T), fmt.Errorf("GetAt: error trying to get. index > Length")
	}
	nodeAtIndex, err := ll.GetNodeAt(index)
	if err != nil {
		return *new(T), err
	}
	return nodeAtIndex.Value, nil
}

func (ll *LinkedList[T]) Forward() iter.Seq2[int, T] {
	var i int
	return func(yield func(int, T) bool) {
		for node := ll.First; node != nil; node = node.Next {
			if !yield(i, node.Value) {
				return
			}
			i++
		}
	}
}

func (ll *LinkedList[T]) Backward() iter.Seq2[int, T] {
	var i int = ll.Length-1
	return func(yield func(int, T) bool) {
		for node := ll.Last; node != nil; node = node.Previous {
			if !yield(i, node.Value) {
				return
			}
			i--
		}
	}
}

func (ll *LinkedList[T]) Print() []T {
	curr := ll.First
	var arr []T
	for curr != nil {
		arr = append(arr, curr.Value)
		curr = curr.Next
	}
	return arr
}
