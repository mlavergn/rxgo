[![Build Status](https://github.com/mlavergn/rxgo/workflows/CI/badge.svg?branch=master)](https://github.com/mlavergn/rxgo/actions)
[![Go Report](https://goreportcard.com/badge/github.com/mlavergn/rxgo)](https://goreportcard.com/report/github.com/mlavergn/rxgo)

# RxGo

ReactiveX inspired lightweight Go channel wrapper

## Note

If you're looking for the full feature set of ReactiveX, it won't be found here. There are existing complete ReactiveX Go implementations that provide that functionality in the style familiar to ReactiveX programmers.

## Usage

Setup the go.mod file as follows:

```text
vi go.mod:
+ require github.com/mlavergn/rxgo v0.20.0
```

Use via import as 'rx':

```golang
package main

import "github.com/mlavergn/rxgo/src/rx"

func main() {
    subject := rx.NewInterval(100)
    observer := rx.NewObserver()
    go func() {
        for {
            select {
            case event := <-observer.Next:
                println(event.(int))
                break
            case err := <-observer.Error:
                println(err)
                return
            case <-observer.Complete:
                return
            }
        }
    }()
    subject.Take(10).Subscribe <- observer
    <-subject.Finalize
}
```

[Playground](https://play.golang.org/p/QNZPDoQAq1j)

## Background

Having working with ReactiveX on other projects, the constructs and patterns ReactiveX defines are a solid blueprint for pub/sub services.

This project arose from the need for a very lightweight ReactiveX feature subset to run on a battery powered device. The existing implementations proved too feature rich for the limitations of the device, and so this project was created.

The result is an API inspired by ReactiveX but which trades conformance for minimal footprint.

## Goals

The following are the project goals:

- Implement Observable / Subject / Subscription patterns
- Model API around on Go channels
- Tighly limit consumption of resources
- gochannel friendly

The current API has ReactiveX naming and behavior, but is firmly centered around Go channels.
