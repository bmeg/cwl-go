
GOPATH := $(shell pwd)
export GOPATH
PATH := ${PATH}:$(shell pwd)/bin
export PATH

cwl-parser:
	go install cwl-parser

cwlgo-tool :
	go install cwlgo-tool

depends :
	go get -d cwlgo-tool

common-workflow-language :
		git clone https://github.com/common-workflow-language/common-workflow-language.git

venv : 
		virtualenv venv
		venv/bin/pip install cwltool

test : cwlgo-tool common-workflow-language venv
	  . venv/bin/activate && cd common-workflow-language && ./run_test.sh RUNNER=../bin/cwlgo-tool common-workflow-language/v1.0/

test_clean : 
		rm -rf venv common-workflow-language