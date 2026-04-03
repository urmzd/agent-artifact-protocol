Create a Python CLI log analyzer tool using argparse.

Include:
- Log parser that handles common formats (Apache, Nginx, JSON structured logs)
- Analysis functions: top IPs, status code distribution, response time percentiles, error rate over time, top endpoints
- Output formatters: table (rich-style ASCII), JSON, CSV
- Main function with argparse: input file, output format, time range filter, top-N limit
- Type hints throughout

Use section IDs: parser, analyzers, formatters, main

Use AAP section markers to delineate each major code block.
Wrap each logical section with `<aap:section id="ID">` and `</aap:section>`.


Output raw code only. No markdown fences, no explanation.