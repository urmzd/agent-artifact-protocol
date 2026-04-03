import json
import sqlite3
import time
import logging
import random
from dataclasses import dataclass
from typing import Optional, List, Dict
import requests
from bs4 import BeautifulSoup
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

@dataclass
class ScraperConfig:
    base_url: str
    rate_limit_seconds: float = 2.0
    max_retries: int = 3
    output_json: str = "products.jsonl"
    db_path: str = "products.db"
    user_agents: List[str] = None

    def __post_init__(self):
        if self.user_agents is None:
            self.user_agents = [
                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"
            ]

class ProductFetcher:
    def __init__(self, config: ScraperConfig):
        self.config = config
        self.session = requests.Session()
        retries = Retry(total=config.max_retries, backoff_factor=1, status_forcelist=[429, 500, 502, 503, 504])
        self.session.mount("https://", HTTPAdapter(max_retries=retries))

    def fetch(self, url: str) -> Optional[str]:
        headers = {'User-Agent': random.choice(self.config.user_agents)}
        try:
            time.sleep(self.config.rate_limit_seconds)
            response = self.session.get(url, headers=headers, timeout=10)
            response.raise_for_status()
            return response.text
        except requests.exceptions.RequestException as e:
            logging.error(f"Failed to fetch {url}: {e}")
            return None

class ProductParser:
    @staticmethod
    def parse(html: str) -> List[Dict]:
        soup = BeautifulSoup(html, 'html.parser')
        products = []
        # Update selectors based on target site structure
        for item in soup.select('.product-item'):
            products.append({
                'name': item.select_one('.title').text.strip() if item.select_one('.title') else None,
                'price': item.select_one('.price').text.strip() if item.select_one('.price') else None,
                'rating': item.select_one('.rating').get('data-value') if item.select_one('.rating') else None,
                'reviews': item.select_one('.review-count').text.strip() if item.select_one('.review-count') else "0",
                'availability': bool(item.select_one('.in-stock')),
                'image_url': item.select_one('img').get('src') if item.select_one('img') else None
            })
        return products

class StorageManager:
    def __init__(self, config: ScraperConfig):
        self.config = config
        self._init_db()

    def _init_db(self):
        with sqlite3.connect(self.config.db_path) as conn:
            conn.execute('''CREATE TABLE IF NOT EXISTS products 
                            (name TEXT, price TEXT, rating TEXT, reviews TEXT, availability BOOLEAN, image_url TEXT)''')

    def save(self, products: List[Dict]):
        with open(self.config.output_json, 'a') as f:
            for p in products:
                f.write(json.dumps(p) + '\n')
        
        with sqlite3.connect(self.config.db_path) as conn:
            conn.executemany("INSERT INTO products VALUES (:name, :price, :rating, :reviews, :availability, :image_url)", products)

def run_scraper(url: str):
    config = ScraperConfig(base_url=url)
    fetcher = ProductFetcher(config)
    parser = ProductParser()
    storage = StorageManager(config)

    html = fetcher.fetch(url)
    if html:
        data = parser.parse(html)
        storage.save(data)
        logging.info(f"Successfully processed {len(data)} products.")

if __name__ == "__main__":
    target = "https://example-ecommerce.com/products"
    run_scraper(target)