package main

import(
  "log"
  "runtime/debug"
)

func HandleErr(err error, msg string) {
  log.Println(msg, ": ", err)
  debug.PrintStack()
}
