use clap::{Parser, ValueEnum};
use csv::{ReaderBuilder, Writer};
use serde::Serialize;
use std::collections::{HashMap, HashSet};
use std::error::Error;
use std::fs::File;
use std::io::{self, BufReader};
use std::path::PathBuf;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum CliError {
    #[error("IO error: {0}")]
    Io(#[from] io::Error),
    #[error("CSV error: {0}")]
    Csv(#[from] csv::Error),
    #[error("Serialization error: {0}")]
    Json(#[from] serde_json::Error),
    #[error("Column not found: {0}")]
    ColumnNotFound(String),
}

#[derive(Copy, Clone, PartialEq, Eq, PartialOrd, Ord, ValueEnum)]
pub enum OutputFormat {
    Table,
    Markdown,
    Json,
    Csv,
}

#[derive(Parser)]
#[command(author, version, about)]
pub struct Args {
    pub input: PathBuf,
    #[arg(short, long, value_enum, default_value = "table")]
    pub format: OutputFormat,
    #[arg(short, long)]
    pub distinct: Option<String>,
    #[arg(short, long)]
    pub limit: Option<usize>,
    #[arg(short = 'n', long, default_value = "10")]
    pub head: usize,
}

pub struct Processor {
    pub headers: Vec<String>,
    pub records: Vec<HashMap<String, String>>,
}

impl Processor {
    pub fn new(path: PathBuf) -> Result<Self, CliError> {
        let file = File::open(path)?;
        let mut reader = ReaderBuilder::new().from_reader(BufReader::new(file));
        let headers: Vec<String> = reader.headers()?.iter().map(|s| s.to_string()).collect();
        let records = reader
            .records()
            .map(|r| headers.iter().cloned().zip(r?.iter().map(|s| s.to_string())).collect())
            .collect::<Result<Vec<_>, _>>()?;
        Ok(Self { headers, records })
    }

    pub fn get_distinct_counts(&self, col: &str) -> Result<Vec<(String, usize)>, CliError> {
        let mut counts = HashMap::new();
        for rec in &self.records {
            let val = rec.get(col).ok_or_else(|| CliError::ColumnNotFound(col.to_string()))?;
            *counts.entry(val.clone()).or_insert(0) += 1;
        }
        let mut sorted: Vec<_> = counts.into_iter().collect();
        sorted.sort_by(|a, b| b.1.cmp(&a.1));
        Ok(sorted)
    }
}

fn main() -> Result<(), Box<dyn Error>> {
    let args = Args::parse();
    let proc = Processor::new(args.input)?;

    if let Some(col) = args.distinct {
        let counts = proc.get_distinct_counts(&col)?;
        let limit = args.limit.unwrap_or(args.head);
        
        println!("| Value | Count |");
        println!("| --- | --- |");
        for (val, count) in counts.into_iter().take(limit) {
            println!("| {} | {} |", val, count);
        }
    } else {
        let limit = args.limit.unwrap_or(args.head);
        let records: Vec<_> = proc.records.into_iter().take(limit).collect();
        match args.format {
            OutputFormat::Json => println!("{}", serde_json::to_string_pretty(&records)?),
            OutputFormat::Csv => {
                let mut wtr = Writer::from_writer(io::stdout());
                for rec in &records { wtr.serialize(rec)?; }
            }
            OutputFormat::Markdown => {
                println!("| {} |", proc.headers.join(" | "));
                println!("| {} |", proc.headers.iter().map(|_| "---").collect::<Vec<_>>().join(" | "));
                for rec in &records {
                    let row: Vec<String> = proc.headers.iter().map(|h| rec.get(h).cloned().unwrap_or_default()).collect();
                    println!("| {} |", row.join(" | "));
                }
            }
            OutputFormat::Table => {
                for rec in &records { println!("{:?}", rec); }
            }
        }
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_distinct_aggregation() {
        let mut rec1 = HashMap::new(); rec1.insert("cat".to_string(), "A".to_string());
        let mut rec2 = HashMap::new(); rec2.insert("cat".to_string(), "A".to_string());
        let mut rec3 = HashMap::new(); rec3.insert("cat".to_string(), "B".to_string());
        let p = Processor { headers: vec!["cat".to_string()], records: vec![rec1, rec2, rec3] };
        
        let counts = p.get_distinct_counts("cat").unwrap();
        assert_eq!(counts[0], ("A".to_string(), 2));
        assert_eq!(counts[1], ("B".to_string(), 1));
    }
}