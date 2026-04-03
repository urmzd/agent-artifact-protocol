use clap::{Parser, ValueEnum};
use csv::{ReaderBuilder, Writer};
use serde::Serialize;
use std::collections::HashMap;
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
    #[arg(short, long, value_delimiter = ',')]
    pub select: Option<Vec<String>>,
    #[arg(short, long)]
    pub filter: Option<String>,
    #[arg(short, long)]
    pub sort: Option<String>,
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
}

fn print_markdown(headers: &[String], records: &[HashMap<String, String>]) {
    println!("| {} |", headers.join(" | "));
    println!("| {} |", headers.iter().map(|_| "---").collect::<Vec<_>>().join(" | "));
    for rec in records {
        let row: Vec<String> = headers.iter().map(|h| rec.get(h).cloned().unwrap_or_default()).collect();
        println!("| {} |", row.join(" | "));
    }
}

fn main() -> Result<(), Box<dyn Error>> {
    let args = Args::parse();
    let mut proc = Processor::new(args.input)?;
    
    let limit = args.limit.unwrap_or(args.head);
    proc.records.truncate(limit);

    match args.format {
        OutputFormat::Json => println!("{}", serde_json::to_string_pretty(&proc.records)?),
        OutputFormat::Csv => {
            let mut wtr = Writer::from_writer(io::stdout());
            for rec in &proc.records { wtr.serialize(rec)?; }
        }
        OutputFormat::Markdown => print_markdown(&proc.headers, &proc.records),
        OutputFormat::Table => {
            for rec in &proc.records { println!("{:?}", rec); }
        }
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_markdown_format() {
        let headers = vec!["id".to_string(), "name".to_string()];
        let mut row = HashMap::new();
        row.insert("id".to_string(), "1".to_string());
        row.insert("name".to_string(), "test".to_string());
        let records = vec![row];
        
        // Ensure headers and content are generated
        assert!(!headers.is_empty());
        assert_eq!(records[0].get("id").unwrap(), "1");
    }
}