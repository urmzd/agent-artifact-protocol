<aap:target id="root">
use clap::{Parser, ValueEnum};
use serde::{Deserialize, Serialize};
use std::path::PathBuf;
use thiserror::Error;

<aap:target id="error-handling">
#[derive(Error, Debug)]
pub enum CliError {
    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),
    #[error("CSV error: {0}")]
    Csv(#[from] csv::Error),
    #[error("Serialization error: {0}")]
    Serialization(#[from] serde_json::Error),
    #[error("Invalid arguments: {0}")]
    InvalidArgs(String),
}
</aap:target>

#[derive(Copy, Clone, PartialEq, Eq, PartialOrd, Ord, ValueEnum)]
pub enum OutputFormat {
    Table,
    Json,
    Csv,
}

<aap:target id="cli-definition">
#[derive(Parser)]
#[command(author, version, about = "CSV Processor")]
pub struct Cli {
    #[arg(short, long)]
    pub input: PathBuf,

    #[arg(short, long, value_enum, default_value_t = OutputFormat::Table)]
    pub format: OutputFormat,

    #[arg(short, long, value_delimiter = ',')]
    pub columns: Option<Vec<String>>,

    #[arg(long)]
    pub filter: Option<String>,

    #[arg(long)]
    pub sort_by: Option<String>,

    #[arg(short, long)]
    pub limit: Option<usize>,
}
</aap:target>

<aap:target id="processor">    pub fn process(args: Cli) -> Result<(), CliError> {
        let mut reader = csv::Reader::from_path(&args.input)?;
        let headers: Vec<String> = reader.headers()?.iter().map(|s| s.to_string()).collect();
        
        let mut records: Vec<serde_json::Value> = reader
            .deserialize()
            .collect::<Result<Vec<serde_json::Value>, _>>()?;

        // Apply column selection
        let columns = args.columns.unwrap_or(headers);
        records = records.into_iter().map(|mut r| {
            let obj = r.as_object_mut().unwrap();
            obj.retain(|k, _| columns.contains(k));
            r
        }).collect();

        // Sorting
        if let Some(col) = args.sort_by {
            records.sort_by(|a, b| {
                a.get(&col).partial_cmp(&b.get(&col)).unwrap_or(std::cmp::Ordering::Equal)
            });
        }

        // Limit
        if let Some(l) = args.limit {
            records.truncate(l);
        }

        // Output
        match args.format {
            OutputFormat::Json => println!("{}", serde_json::to_string_pretty(&records)?),
            OutputFormat::Csv => {
                let mut wtr = csv::Writer::from_writer(std::io::stdout());
                wtr.write_record(&columns)?;
                for rec in records {
                    let row: Vec<String> = columns.iter().map(|c| rec[c].as_str().unwrap_or("").to_string()).collect();
                    wtr.write_record(&row)?;
                }
                wtr.flush()?;
            }
            OutputFormat::Table => {
                println!("| {} |", columns.join(" | "));
                println!("| {} |", columns.iter().map(|_| "---").collect::<Vec<_>>().join(" | "));
                for rec in records {
                    let row: Vec<String> = columns.iter().map(|c| rec[c].as_str().unwrap_or("").to_string()).collect();
                    println!("| {} |", row.join(" | "));
                }
            }
        }

        Ok(())
    }
</aap:target>

<aap:target id="tests">
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parsing() {
        let cli = Cli::try_parse_from(&["bin", "-i", "test.csv"]);
        assert!(cli.is_ok());
    }
}
</aap:target>

fn main() -> Result<(), CliError> {
    let args = Cli::parse();
    Processor::process(args)
}
</aap:target>