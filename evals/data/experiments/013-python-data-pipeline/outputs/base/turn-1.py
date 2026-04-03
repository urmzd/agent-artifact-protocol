def transform(self):
        # Existing logic
        self.df['profit_margin'] = (self.df['revenue'] - self.df['cost']) / self.df['revenue']
        
        bins = [-float('inf'), 500, 2000, float('inf')]
        labels = ['Low-Value', 'Mid-Value', 'High-Value']
        self.df['product_tier'] = pd.cut(self.df['revenue'], bins=bins, labels=labels)
        
        # New transformation: Customer Lifetime Value (CLV)
        clv_map = self.df.groupby('customer_id')['revenue'].sum().rename('customer_lifetime_value')
        self.df = self.df.merge(clv_map, on='customer_id', how='left')
        
        # Aggregate by region
        self.regional_summary = self.df.groupby('region').agg({
            'revenue': 'sum',
            'profit_margin': 'mean'
        }).reset_index()