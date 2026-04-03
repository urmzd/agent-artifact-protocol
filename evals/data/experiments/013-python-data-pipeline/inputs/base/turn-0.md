Create a Python ETL data pipeline for processing CSV sales data.

Include:
- Extraction: read CSV files, handle encoding issues, parse dates
- Transformation: clean column names, calculate derived metrics (profit margin, YoY growth), categorize products, aggregate by region
- Validation: schema validation with expected types, null checks, range checks, duplicate detection
- Loading: write to Parquet and JSON, generate summary statistics
- Use dataclasses for pipeline config, pandas for data processing

Use section IDs: extraction, transformation, validation, loading

Use AAP section markers to delineate each major code block.
Wrap each logical section with `<aap:section id="ID">` and `</aap:section>`.


Output raw code only. No markdown fences, no explanation.