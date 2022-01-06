#!/bin/bash

sed -i "s|__DataId__|${dataId}|g" /etc/confd/templates/constants.template
sed -i "s|__DataId__|${dataId}|g" /etc/confd/conf.d/template.toml

confd ${watch} -interval ${interval} -backend nacosv2 -log-level ${loglevel} -scheme ${scheme} -host ${host} -port ${port} -uri ${uri} -group ${group} -namespace ${namespace}

