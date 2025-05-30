#!/bin/sh

OUTPUT_DIR="./report/report_$(date --iso-8601=seconds)"

jmeter -n -t config/jmeter/load_test.jmx -l "$OUTPUT_DIR/result.jtl" -e -o "$OUTPUT_DIR"
