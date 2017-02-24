#!/usr/bin/env python

import sys
import json
import cwllite
import yaml

wf = cwllite.load_proto(sys.argv[1])

print yaml.safe_dump(cwllite.to_cwl(wf), default_flow_style=False)
