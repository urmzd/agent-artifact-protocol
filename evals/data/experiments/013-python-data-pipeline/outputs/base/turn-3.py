def load(self):
        Path(self.config.output_dir).mkdir(parents=True, exist_ok=True)
        
        # Save Parquet
        self.df.to_parquet(f"{self.config.output_dir}/processed_sales.parquet")
        
        # Calculate Regional Summary Statistics
        regional_stats = self.df.groupby('region').agg(
            total_revenue=('revenue', 'sum'),
            order_count=('order_id', 'count'),
            avg_order_value=('revenue', 'mean')
        ).reset_index()
        
        # Save Summary CSV
        regional_stats.to_csv(f"{self.config.output_dir}/regional_summary.csv", index=False)
        
        # Save JSON Summary
        summary = {
            "total_records": int(len(self.df)),
            "avg_revenue": float(self.df['revenue'].mean()),
            "regional_performance": regional_stats.to_dict(orient='records')
        }
        with open(f"{self.config.output_dir}/summary.json", 'w') as f:
            json.dump(summary, f, indent=4)