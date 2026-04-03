<aap:target id="etl-pipeline">
import pandas as pd
import numpy as np
from dataclasses import dataclass
from typing import List, Dict
import logging

<aap:target id="config-settings">
@dataclass
class ETLConfig:
    input_path: str
    output_parquet: str
    output_json: str
    expected_columns: List[str]
    date_col: str
</aap:target>

class SalesETLPipeline:
    def __init__(self, config: ETLConfig):
        self.config = config
        self.logger = logging.getLogger(__name__)

    <aap:target id="extraction-logic">
    def extract(self) -> pd.DataFrame:
        self.logger.info("Extracting data...")
        df = pd.read_csv(
            self.config.input_path, 
            encoding='utf-8-sig',
            parse_dates=[self.config.date_col]
        )
        return df
    </aap:target>

    <aap:target id="transformation-logic">        # Aggregation
        region_stats = df.groupby('region').agg({'revenue': 'sum', 'profit_margin': 'mean'})
        
        # Customer lifetime value
        df['customer_lifetime_value'] = df.groupby('customer_id')['revenue'].transform('sum')
        
        return df
</aap:target>

    <aap:target id="validation-logic">
    def validate(self, df: pd.DataFrame):
        # Null checks
        if df.isnull().values.any():
            raise ValueError("Data contains null values")
            
        # Range checks
        if (df['revenue'] < 0).any():
            raise ValueError("Negative revenue detected")
            
        # Duplicate detection
        if df.duplicated().any():
            self.logger.warning("Duplicates found, removing...")
            df.drop_duplicates(inplace=True)
    </aap:target>

    <aap:target id="loading-logic">
    def load(self, df: pd.DataFrame):
        df.to_parquet(self.config.output_parquet)
        df.to_json(self.config.output_json, orient='records')
        
        summary = {
            "total_records": len(df),
            "mean_revenue": float(df['revenue'].mean())
        }
        return summary
    </aap:target>

def run_pipeline(config: ETLConfig):
    pipeline = SalesETLPipeline(config)
    data = pipeline.extract()
    pipeline.validate(data)
    transformed_data = pipeline.transform(data)
    stats = pipeline.load(transformed_data)
    return stats

</aap:target>