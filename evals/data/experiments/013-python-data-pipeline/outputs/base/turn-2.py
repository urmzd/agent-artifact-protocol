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

        # New: Future date validation
        current_date = pd.Timestamp.now()
        for col in self.config.date_columns:
            if (self.df[col] > current_date).any():
                future_dates = self.df[self.df[col] > current_date]
                raise ValueError(f"Future dates detected in {col}: \n{future_dates}")
        
        print("Validation passed.")