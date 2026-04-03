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
    pub records: Vec<HashMap<String, String>>,
}

impl Processor {
    pub fn new(path: PathBuf) -> Result<Self, CliError> {
        let file = File::open(path)?;
        let mut reader = ReaderBuilder::new().from_reader(BufReader::new(file));
        let headers: Vec<String> = reader.headers()?.iter().map(|s| s.to_string()).collect();
        let records = reader
            .records()
            .map(|r| {
                headers.iter().cloned().zip(r?.iter().map(|s| s.to_string())).collect()
            })
            .collect::<Result<Vec<_>, _>>()?;
        Ok(Self { records })
    }

    pub fn filter(&mut self, expr: &str) {
        if let Some(pos) = expr.find("==") {
            let (key, val) = (expr[..pos].trim(), expr[pos+2..].trim());
            self.records.retain(|r| r.get(key).map(|v| v == val).unwrap_or(false));
        }
    }

    pub fn sort(&mut self, column: &str) {
        self.records.sort_by(|a, b| {
            a.get(column).unwrap_or(&"".to_string())
                .cmp(b.get(column).unwrap_or(&"".to_string()))
        });
    }
}

fn main() -> Result<(), Box<dyn Error>> {
    let args = Args::parse();
    let mut proc = Processor::new(args.input)?;

    if let Some(f) = args.filter { proc.filter(&f); }
    if let Some(s) = args.sort { proc.sort(&s); }
    
    let limit = args.limit.unwrap_or(args.head);
    proc.records.truncate(limit);

    match args.format {
        OutputFormat::Json => println!("{}", serde_json::to_string_pretty(&proc.records)?),
        OutputFormat::Csv => {
            let mut wtr = Writer::from_writer(io::stdout());
            for rec in &proc.records { wtr.serialize(rec)?; }
        }
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
    fn test_head_logic() {
        let mut p = Processor { records: vec![HashMap::new(); 20] };
        let head = 5;
        p.records.truncate(head);
        assert_eq!(p.records.len(), 5);
    }
}