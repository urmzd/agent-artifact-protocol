import pandas as pd
import numpy as np
import json
from dataclasses import dataclass
from pathlib import Path
from typing import List, Dict, Any

@dataclass
class ETLConfig:
    input_path: str
    output_dir: str
    date_columns: List[str]
    required_columns: List[str]
    numeric_columns: List[str]

class SalesPipeline:
    def __init__(self, config: ETLConfig):
        self.config = config
        self.df = None

    def extract(self):
        try:
            self.df = pd.read_csv(
                self.config.input_path, 
                encoding='utf-8-sig', 
                parse_dates=self.config.date_columns
            )
            self.df.columns = [c.strip().lower().replace(' ', '_') for c in self.df.columns]
        except Exception as e:
            raise IOError(f"Extraction failed: {e}")

    def validate(self):
        # Null checks
        if self.df[self.config.required_columns].isnull().any().any():
            raise ValueError("Null values found in required columns")
        
        # Duplicate detection
        if self.df.duplicated().any():
            self.df.drop_duplicates(inplace=True)
            
        # Range/Type checks
        for col in self.config.numeric_columns:
            if (self.df[col] < 0).any():
                raise ValueError(f"Negative values found in {col}")
        
        print("Validation passed.")

    def transform(self):
        # Profit Margin
        self.df['profit_margin'] = (self.df['revenue'] - self.df['cost']) / self.df['revenue']
        
        # Category mapping
        bins = [-float('inf'), 500, 2000, float('inf')]
        labels = ['Low-Value', 'Mid-Value', 'High-Value']
        self.df['product_tier'] = pd.cut(self.df['revenue'], bins=bins, labels=labels)
        
        # Aggregate by region
        self.regional_summary = self.df.groupby('region').agg({
            'revenue': 'sum',
            'profit_margin': 'mean'
        }).reset_index()

    def load(self):
        Path(self.config.output_dir).mkdir(parents=True, exist_ok=True)
        
        # Save Parquet
        self.df.to_parquet(f"{self.config.output_dir}/processed_sales.parquet")
        
        # Save JSON Summary
        summary = {
            "total_records": int(len(self.df)),
            "avg_revenue": float(self.df['revenue'].mean()),
            "regional_performance": self.regional_summary.to_dict(orient='records')
        }
        with open(f"{self.config.output_dir}/summary.json", 'w') as f:
            json.dump(summary, f, indent=4)

    def run(self):
        self.extract()
        self.validate()
        self.transform()
        self.load()
        print("Pipeline execution complete.")

if __name__ == "__main__":
    config = ETLConfig(
        input_path="sales_data.csv",
        output_dir="output",
        date_columns=["sale_date"],
        required_columns=["order_id", "revenue", "cost", "region"],
        numeric_columns=["revenue", "cost"]
    )
    
    pipeline = SalesPipeline(config)
    pipeline.run()