
import yaml
from copy import deepcopy
from google.protobuf.json_format import ParseDict, MessageToDict
from .cwl_pb2 import CommandLineTool

def fields_dict2list(doc, *args):
    out = {}
    for k,v in doc.items():
        if k in args:
            if isinstance(v, dict):
                nv = []
                for ek, ev in v.items():
                    i = deepcopy(ev)
                    i['id'] = ek
                    nv.append(i)
                out[k] = nv
            else:
                out[k] = v
        else:
            out[k] = v
    return out

def fields_forcelist(doc, *args):
    out = {}
    for k,v in doc.items():
        if k in args:
            if not isinstance(v, list):
                out[k] = [v]
            else:
                out[k] = v
        else:
            out[k] = v
    return out

def prep_TypeRecord(doc):
    if isinstance(doc, basestring):
        return {"name" : doc}
    if isinstance(doc, dict):
        if doc['type'] == "array":
            return { "items" : {"type" : prep_TypeRecord(doc['items']) } }
    if isinstance(doc, list):
        t = []
        for i in doc:
            t.append(prep_TypeRecord(i))
        return {"oneof" : {"types" : t}}
    return doc

def prep_InputRecordField(doc):
    doc = fields_forcelist(doc, "doc")
    doc['type'] = prep_TypeRecord(doc['type'])
    return doc

def prep_OutputRecordField(doc):
    doc = fields_forcelist(doc, "doc")
    doc['type'] = prep_TypeRecord(doc['type'])
    if 'outputBinding' in doc:
        doc['outputBinding'] = prep_CommandOutputBinding(doc['outputBinding'])
    return doc

def prep_CommandOutputBinding(doc):
    doc = fields_forcelist(doc, "glob")
    return doc

def prep_InputRecordField_list(doc):
    out = []
    for i in doc:
        out.append(prep_InputRecordField(i))
    return out

def prep_OutputRecordField_list(doc):
    out = []
    for i in doc:
        out.append(prep_OutputRecordField(i))
    return out

def prep_CommandLineTool(doc):
    doc = fields_dict2list(doc, "inputs", "outputs", "hints")
    doc = fields_forcelist(doc, "baseCommand", "doc")
    doc['inputs'] = prep_InputRecordField_list(doc['inputs'])
    doc['outputs'] = prep_OutputRecordField_list(doc['outputs'])
    return doc

MUTATORS = {
    "CommandLineTool" : prep_CommandLineTool
}

def load(path):
    with open(path) as handle:
        data = handle.read()
        doc = yaml.load(data)

    if doc['class'] == "CommandLineTool":
        doc = prep_CommandLineTool(doc)
        out = CommandLineTool()
        print doc
        ParseDict(doc, out)
    return out

def to_dict(pb):
    return MessageToDict(pb)
