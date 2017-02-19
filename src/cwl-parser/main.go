package main

import (
	"cwlparser"
	"flag"
	"fmt"
	//"io/ioutil"
	//"log"
	//"os"
	//"path/filepath"
	//"strings"
	//"time"
)

func main() {
	//var version_flag = flag.Bool("version", false, "version")
	//var tmp_outdir_prefix_flag = flag.String("tmp-outdir-prefix", "./", "Temp output prefix")
	//var tmpdir_prefix_flag = flag.String("tmpdir-prefix", "/tmp", "Tempdir prefix")
	//var outdir = flag.String("outdir", "./", "Outdir")
	//var quiet_flag = flag.Bool("quiet", false, "quiet")
	flag.Parse()
  
  cwl_path := flag.Arg(0)

  parser := cwlparser.NewParser(cwl_path)
  
  d, _ := parser.Parse()
  
  fmt.Printf("%#v\n", d)
  
}