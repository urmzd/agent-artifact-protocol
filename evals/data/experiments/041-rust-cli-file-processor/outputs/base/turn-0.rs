use clap::{Parser, ValueEnum};
use csv::{ReaderBuilder, Writer};
use serde::{Deserialize, Serialize};
use serde_json::Value;
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
    #[error("Invalid column: {0}")]
    InvalidColumn(String),
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
}

pub struct Processor {
    headers: Vec<String>,
    records: Vec<HashMap<String, String>>,
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
        Ok(Self { headers, records })
    }

    pub fn filter(&mut self, expr: &str) {
        let parts: Vec<&str> = expr.split_operator().collect();
        if parts.len() == 3 {
            let (col, op, val) = (parts[0], parts[1], parts[2]);
            self.records.retain(|rec| {
                let cell = rec.get(col).map(|s| s.as_str()).unwrap_or("");
                match op {
                    "==" => cell == val,
                    ">" => cell.parse::<f64>().unwrap_or(0.0) > val.parse::<f64>().unwrap_or(0.0),
                    _ => true,
                }
            });
        }
    }

    pub fn sort(&mut self, column: &str) {
        self.records.sort_by(|a, b| {
            a.get(column).unwrap_or(&"".to_string())
                .partial_cmp(b.get(column).unwrap_or(&"".to_string()))
                .unwrap()
        });
    }
}

trait SplitOp { fn split_operator(&self) -> Vec<&str>; }
impl SplitOp for str {
    fn split_operator(&self) -> Vec<&str> {
        if self.contains("==") { self.split("==").collect() }
        else if self.contains(">") { self.split(">").collect() }
        else { vec![] }
    }
}

fn main() -> Result<(), Box<dyn Error>> {
    let args = Args::parse();
    let mut proc = Processor::new(args.input)?;

    if let Some(f) = args.filter { proc.filter(&f); }
    if let Some(s) = args.sort { proc.sort(&s); }
    if let Some(l) = args.limit { proc.records.truncate(l); }

    match args.format {
        OutputFormat::Json => {
            println!("{}", serde_json::to_string_pretty(&proc.records)?);
        }
        OutputFormat::Csv => {
            let mut wtr = Writer::from_writer(io::stdout());
            for rec in &proc.records {
                wtr.serialize(rec)?;
            }
        }
        OutputFormat::Table => {
            for rec in &proc.records {
                println!("{:?}", rec);
            }
        }
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_filter_logic() {
        let mut records = vec![
            HashMap::from([("age".to_string(), "20".to_string())]),
            HashMap::from([("age".to_string(), "30".to_string())]),
        ];
        let mut p = Processor { headers: vec!["age".to_string()], records };
        p.filter("age>25");
        assert_eq!(p.records.len(), 1);
        assert_eq!(p.records[0].get("age").unwrap(), "30");
    }
}