

Build Avro Based code
---------------------
```
git clone https://github.com/common-workflow-language/common-workflow-language.git
virtualenv venv
. venv/bin/activate
pip install schema-salad
python -mschema_salad --print-avro ./common-workflow-language/v1.0/CommonWorkflowLanguage.yml > cwl.avsc
./tools/cwl-avro-go.py cwl.avsc > cwl.go
```
