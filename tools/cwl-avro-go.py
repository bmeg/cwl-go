#!/usr/bin/env python

import re
import sys
import json
import jinja2

SKIP = [
    "Any"
]

RECORD_MAP_TEMPLATE = """
var TYPES = map[string]reflect.Type {
{%- for name in names %}
    "{{name}}" : reflect.TypeOf({{name}}{}),
{%- endfor %}
}
"""

RECORD_TEMPLATE = """
type {{name}} struct {
{%- for field in fields %}
    {{ field }} {{ fields[field].type}}   {%if field.array %}[]{% endif %}`json:"{{ fields[field].name}}"`
{%- endfor %}
}
{% if hasid %}
func (self *{{name}}) GetID() string {
    return self.Id
}
{% endif %}
"""

UNION_TEMPLATE = """

type {{name}} struct {
{%- for t in types %}
    {{t.name}} {%if t.array %}[]{% else %}*{% endif %}{{t.type}}
{%- endfor %}
}

"""

ENUM_TEMPLATE = """
type {{name}} string
const (
{%- for symbol in symbols %}
    {{name}}_{{symbol}} {{name}} = "{{symbols[symbol]}}"
{%- endfor %}
)

"""

TYPE_MAPPING = {
    "Any" : "interface{}",
    "boolean" : "bool",
    "long" : "int64",
}

def fixCaps(s):
    s = re.sub(r"[\-\.]", "_", s)
    return s[0:1].capitalize() + s[1:]


class Union:
    def __init__(self, name):
        self.name = name
        self.types = []

    def add_type(self, typeName, array):
        self.types.append( {"name" : fixCaps(typeName) + "Value", "type" : typeName, "array" : array} )

class Field:
    def __init__(self, type, name):
        self.type = type
        self.name = name

class RecordField:
    def __init__(self, type, name):
        self.type = name
        self.type_list = type

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
                fields[fixCaps(k['name'])] = Field(type=t, name=k['name'] )
            else:
                if isinstance(t, basestring):
                    if k['name'] != "type":
                        fields[fixCaps(k['name'])] = Field(type=t, name=k["name"] )
                elif isinstance(t, dict) and t['type'] == "array" and isinstance(t['items'], basestring):
                    l = t['items']
                    if l in TYPE_MAPPING:
                        l = TYPE_MAPPING[l]
                    fields[fixCaps(k['name'])] = Field(type="[]%s" % (l), name=k['name'])
                elif isinstance(t, list):
                    fields[fixCaps(k['name'])] = RecordField(type=t, name="%s%s" % (self.doc['name'], fixCaps(k['name'])) )
                    pass
                elif isinstance(t, dict) and t['type'] == "array" and isinstance(t['items'], list):
                    pass
                elif isinstance(t, dict) and t['type'] == "record":
                    fields[fixCaps(k['name'])] = Field(type=t['name'], name=k['name'])
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
        fields = {}

        has_id = False
        for k,v in self.get_fields().items():
            fields[k] = v
            if k == "Id":
                has_id = True

        kwargs = {
            "name" : self.doc['name'],
            "fields" : fields,
            "hasid" : has_id
        }
        return jinja2.Template(RECORD_TEMPLATE).render(**kwargs)



class Schema:
    def __init__(self, doc):
        self.doc = doc
        self._records = self.list_records()

    def generate(self):
        out = """
package cwlparser

import (
  "reflect"
)

"""
        for k, v in self.list_enums().items():
            if k not in SKIP:
                out += self.gen_enum(v)

        """
        for k, v in self.list_union().items():
            out += self.gen_union(v)
        """

        records = self.list_records()
        out += self.gen_record_map(records)

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

    """
    def list_union(self):
        out = {}
        for k, v in self._records.items():
            for f in v.get_fields().values():
                if isinstance(f['type'], list):
                    t = f['type']
                    if "null" in t:
                        t.remove("null")
                    if len(t) > 1:
                        u = Union( "%s%s" % (k, fixCaps(f['name'])) )
                        for i in t:
                            if isinstance(i,basestring):
                                u.add_type( i, False )
                            elif isinstance(i,dict):
                                if i['type'] == "record":
                                    u.add_type(i['name'], False)
                                elif i['type'] == "array":
                                    if isinstance(i['items'], list):
                                        uc = Union( "%s%sElement" % (k, fixCaps(f['name'])) )
                                        for ut in i['items']:
                                            uc.add_type( ut, False)
                                        out[uc.name] = uc
                                        u.add_type(uc.name, True)
                                    else:
                                        u.add_type(i['items'], True)
                                elif i['type'] == "enum":
                                    pass
                                else:
                                    print "missed element", i
                            else:
                                print "missed", i
                        out[u.name] = u
        return out
    """

    def scan_record_list(self, record):
        out = {}
        if record['type'] == "record":
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

    def gen_enum(self, record):
        symbols = {}
        for s in record["symbols"]:
            symbols[ fixCaps(s) ] = s
        kwargs = {
            "name" : record["name"],
            "symbols" : symbols
        }
        return jinja2.Template(ENUM_TEMPLATE).render(**kwargs)

    def gen_union(self, union):
        types = []
        for t in union.types:
            #if there is a single and array version of the same value type
            #then skip the singleton
            if not t['array']:
                found = False
                for o in union.types:
                    if o['type'] == t['type'] and o['array']:
                        found = True
                if not found:
                    types.append(t)
            else:
                types.append(t)


        kwargs = {
            "name" : union.name,
            "types" : types
        }
        return jinja2.Template(UNION_TEMPLATE).render(**kwargs)

    def gen_record_map(self, records):
        kwargs = {
            "names" : records.keys()
        }
        return jinja2.Template(RECORD_MAP_TEMPLATE).render(**kwargs)




if __name__ == "__main__":

    with open(sys.argv[1]) as handle:
        doc = json.loads(handle.read())
        schema = Schema(doc)

    src = schema.generate()
    print src
