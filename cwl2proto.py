#!/usr/bin/env python

import sys
import json
import cwllite
import yaml

wf = cwllite.load_cwl(sys.argv[1], True)

print yaml.safe_dump(cwllite.to_dict(wf), default_flow_style=False)
