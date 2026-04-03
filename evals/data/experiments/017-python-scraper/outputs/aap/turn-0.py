import json
import sqlite3
import time
import random
import logging
from dataclasses import dataclass
from typing import Optional, Dict, Any
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry
from bs4 import BeautifulSoup

<aap:target id="scraper-module">
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

<aap:target id="config-block">
@dataclass
class ScraperConfig:
    base_url: str = "https://example-ecommerce.com"
    rate_limit: float = 1.5
    max_retries: int = 3
    backoff_factor: float = 0.5
    output_jsonl: str = "products.jsonl"
    db_path: str = "products.db"
    user_agents: list = None

    def __post_init__(self):
        if self.user_agents is None:
            self.user_agents = [
                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"
            ]
</aap:target>

<aap:target id="fetcher-block">
class ProductFetcher:
    def __init__(self, config: ScraperConfig):
        self.config = config
        self.session = requests.Session()
        retries = Retry(
            total=config.max_retries,
            backoff_factor=config.backoff_factor,
            status_forcelist=[429, 500, 502, 503, 504]
        )
        self.session.mount("https://", HTTPAdapter(max_retries=retries))

    def fetch(self, url: str) -> Optional[str]:
        headers = {"User-Agent": random.choice(self.config.user_agents)}
        try:
            time.sleep(self.config.rate_limit)
            response = self.session.get(url, headers=headers, timeout=10)
            response.raise_for_status()
            return response.text
        except requests.exceptions.RequestException as e:
            logger.error(f"Failed to fetch {url}: {e}")
            return None
</aap:target>

<aap:target id="parser-block">
def parse_product(html: str) -> Dict[str, Any]:
    soup = BeautifulSoup(html, 'html.parser')
    return {
        "name": soup.select_one(".product-title").text.strip() if soup.select_one(".product-title") else None,
        "price": soup.select_one(".price").text.strip() if soup.select_one(".price") else None,
        "rating": soup.select_one(".rating").get("data-value") if soup.select_one(".rating") else 0,
        "review_count": soup.select_one(".review-count").text.strip() if soup.select_one(".review-count") else 0,
        "availability": bool(soup.select_one(".in-stock")),
        "image_url": soup.select_one(".main-image").get("src") if soup.select_one(".main-image") else None
    }
</aap:target>

<aap:target id="storage-block">
class DataStorage:
    def __init__(self, config: ScraperConfig):
        self.config = config
        self._init_db()

    def _init_db(self):
        with sqlite3.connect(self.config.db_path) as conn:
            conn.execute("""
                CREATE TABLE IF NOT EXISTS products 
                (name TEXT, price TEXT, rating REAL, review_count INTEGER, availability BOOLEAN, image_url TEXT)
            """)

    def save(self, data: Dict[str, Any]):
        with open(self.config.output_jsonl, "a") as f:
            f.write(json.dumps(data) + "\n")
        
        with sqlite3.connect(self.config.db_path) as conn:
            conn.execute("INSERT INTO products VALUES (?, ?, ?, ?, ?, ?)", tuple(data.values()))
</aap:target>
</aap:target>