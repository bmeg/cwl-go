#!/usr/bin/env python

import re
import sys
import json
import jinja2

SKIP = [
    "Any",
    "ArraySchema",
    "CommandInputArraySchema",
    "OutputArraySchema",
    "InputArraySchema",
    "CommandOutputArraySchema",
    "SchemaDefRequirement"
]

MESSAGE_TEMPLATE = """
message {{name}} {
{%- for field in fields %}
    {{ field.type}} {{ field.name }} = {{field.num}}; {% if field.comment %}//{{field.comment}}{% endif %}
{%- endfor %}
}

"""


TYPE_MAPPING = {
    "Any" : "google.protobuf.Struct",
    "boolean" : "bool",
    "long" : "int64",
    "int" : "int64",
    "CWLVersion" : "string",
    "LinkMergeMethod" : "string",
    "ScatterMethod" : "string"
}

FIELD_FIX = {
    "Workflow" : {
        "requirements" : "repeated google.protobuf.Struct"
    },
    "WorkflowStepInput" : {
        "source" : "repeated string"
    },
    "WorkflowOutputParameter" : {
        "outputSource" : "repeated string"
    },
    "ExpressionTool" : {
        "expression" : "string",
        "requirements" : "repeated google.protobuf.Struct"
    },
    "CommandOutputBinding" : {
        "glob" : "repeated string",
        "outputEval" : "string"
    },
    "CommandLineTool" : {
        "stdout" : "string",
        "stdin" : "string",
        "stderr" : "string",
        "baseCommand" : "repeated string",
        "requirements" : "repeated google.protobuf.Struct",
        "arguments" : "repeated CommandLineBinding"
    },
    "CommandInputParameter" : {
        "default" : "DataRecord"
    },
    "WorkflowStep" : {
        "run" : "RunRecord",
        "scatter" : "repeated string",
        "requirements" : "repeated google.protobuf.Struct",
        "out" : "repeated WorkflowStepOutput"
    },
    "EnvironmentDef" : {
        "envValue" : "string"
    },
    "*" : {
        "doc" : "repeated string",
        "type" : "TypeRecord",
        "valueFrom" : "string",
        "format" : "repeated string",
        "secondaryFiles" : "repeated string"
    }
}

def fixCaps(s):
    s = re.sub(r"[\-\.]", "_", s)
    return s[0:1].capitalize() + s[1:]


class Field:
    def __init__(self, type, name, comment = None):
        self.type = type
        self.name = name
        self.comment = comment

class Record:
    def __init__(self, doc):
        self.doc = doc

    def get_type(self):
        return self.doc['type']

    def get_fields(self):
        fields = {}
        for k in self.doc['fields']:
            t = k['type']
            if isinstance(k['type'], list):
                if "null" in t:
                    t.remove("null")
                if len(t) == 1:
                    t = t[0]
            if t in ["string", "boolean", "int", "long", "float", "double", "Any"]:
                if t in TYPE_MAPPING:
                    t = TYPE_MAPPING[t]
                if self.doc['name'] in FIELD_FIX and k['name'] in FIELD_FIX[self.doc['name']]:
                    t = FIELD_FIX[self.doc['name']][k['name']]
                elif k['name'] in FIELD_FIX["*"]:
                    t = FIELD_FIX["*"][k['name']]
                fields[k['name']] = Field(type=t, name=k['name'] )
            else:
                if isinstance(t, basestring):
                    if k['name'] != "type":
                        if t in TYPE_MAPPING:
                            t = TYPE_MAPPING[t]
                        fields[k['name']] = Field(type=t, name=k["name"] )
                elif isinstance(t, dict) and t['type'] == "array" and isinstance(t['items'], basestring):
                    l = t['items']
                    if l in TYPE_MAPPING:
                        l = TYPE_MAPPING[l]
                    fields[k['name']] = Field(type="repeated %s" % (l), name=k['name'])
                elif isinstance(t, list):
                    tname = "Any"
                    if self.doc['name'] in FIELD_FIX and k['name'] in FIELD_FIX[self.doc['name']]:
                        tname = FIELD_FIX[self.doc['name']][k['name']]
                    elif k['name'] in FIELD_FIX["*"]:
                        tname = FIELD_FIX["*"][k['name']]
                    fields[k['name']] = Field(type=tname, name=k['name'], comment=str(t))
                    pass
                elif isinstance(t, dict) and t['type'] == "array" and isinstance(t['items'], list):
                    tname = "Any"
                    if self.doc['name'] in FIELD_FIX and k['name'] in FIELD_FIX[self.doc['name']]:
                        tname = FIELD_FIX[self.doc['name']][k['name']]
                    elif k['name'] in FIELD_FIX["*"]:
                        tname = FIELD_FIX["*"][k['name']]
                    fields[k['name']] = Field(type=tname, name=k['name'], comment=str(t))
                elif isinstance(t, dict) and t['type'] == "record":
                    fields[k['name']] = Field(type=t['name'], name=k['name'])
                elif isinstance(t, dict) and t['type'] == "enum":
                    pass
                else:
                    print "missing", t
        return fields

    def extend(self, records):
        if 'extends' in self.doc:
            e = self.doc['extends']
            if not isinstance(e, list):
                e = [e]
            for el in e:
                ename = el.split("#")[1]
                if ename in records:
                    #sys.stderr.write("extends!!! %s" % ename)
                    for v in records[ename].doc['fields']:
                        self.doc['fields'].append(v)
                    #sys.stderr.write("After:%s\n" % self.doc['fields'])

    def render(self):
        fields = []

        i = 1;
        for k,v in self.get_fields().items():
            fields.append(v)
            v.num = i
            i += 1
        kwargs = {
            "name" : self.doc['name'],
            "fields" : fields,
        }
        return jinja2.Template(MESSAGE_TEMPLATE).render(**kwargs)



class Schema:
    def __init__(self, doc):
        self.doc = doc
        self._records = self.list_records()

    def generate(self):
        out = """
syntax = "proto3";

import "google/protobuf/struct.proto";

message ArrayRecord {
    TypeRecord items = 1;
}

message FieldRecord {
    string name = 1;
    TypeRecord type = 2;
}

message RecordRecord {
    string name = 1;
    repeated FieldRecord fields = 2;
}

message EnumRecord {
    string name = 1;
    repeated string symbols = 2;
}

message OneOfRecord {
    repeated TypeRecord types = 1;
}

message TypeRecord {
    oneof type {
        string name = 1;
        ArrayRecord array = 2;
        OneOfRecord oneof = 3;
        RecordRecord record = 4;
        EnumRecord enum = 5;
    }
}

message RunRecord {
    oneof run {
        string path = 1;
        CommandLineTool commandline = 2;
        ExpressionTool expression = 3;
        Workflow workflow = 4;
    }
}

message DataRecord {
    oneof data {
        string string_value = 1;
        google.protobuf.Struct struct_value = 2;
        double float_value = 3;
        int64 int_value = 4;
        google.protobuf.ListValue list_value = 5;
        bool bool_value = 6;
    }
}

message CWLClass {
    oneof class {
        Workflow workflow = 1;
        CommandLineTool commandline = 2;
        ExpressionTool expression = 3;
    }
}

message GraphRecord {
    string cwlVersion = 1;
    repeated CWLClass graph = 2;
}
"""
        records = self.list_records()
        for k, v in records.items():
            out += v.render()
        return out

    def list_records(self):
        out = {}
        for i in self.doc:
            for k, v in self.scan_record_list(i).items():
                out[k] = v
        for v in out.values():
            v.extend(out)
        return out

    def list_enums(self):
        out = {}
        for i in self.doc:
            if i['type'] == "enum":
                out[i['name']] = i
        return out


    def scan_record_list(self, record):
        out = {}
        if record['type'] == "record":
            if record["name"] not in SKIP:
                out[record["name"]] = Record(record)
                for f in record['fields']:
                    for k, v in self.scan_record_list(f):
                        out[k] = v
                    if isinstance(f['type'], list):
                        for c in f['type']:
                            if isinstance(c, dict):
                                for k,v in self.scan_record_list(c).items():
                                    out[k] = v
        return out


if __name__ == "__main__":

    with open(sys.argv[1]) as handle:
        doc = json.loads(handle.read())
        schema = Schema(doc)

    src = schema.generate()
    print src
