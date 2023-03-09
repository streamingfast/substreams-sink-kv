package main

import (
	_ "github.com/streamingfast/kvdb/store/badger"
	_ "github.com/streamingfast/kvdb/store/badger3"
	_ "github.com/streamingfast/kvdb/store/bigkv"
	_ "github.com/streamingfast/kvdb/store/netkv"
	_ "github.com/streamingfast/kvdb/store/tikv"
)

func main() {
	if err := rootCmd.Execute(); err != nil {

	}
}
