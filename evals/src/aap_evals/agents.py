"""Artifact generation and LLM model factory."""

from __future__ import annotations

import re

from pydantic_ai import Agent
from pydantic_ai.models import Model
from pydantic_ai.models.fallback import FallbackModel
from pydantic_ai.models.google import GoogleModel
from pydantic_ai.models.openai import OpenAIChatModel
from pydantic_ai.providers.google import GoogleProvider
from pydantic_ai.providers.ollama import OllamaProvider
from pydantic_ai.providers.openai import OpenAIProvider


# ── Provider defaults ─────────────────────────────────────────────────────

PROVIDER_DEFAULTS: dict[str, str] = {
    "google": "gemini-3.1-flash-lite-preview",
    "openai": "gpt-4o-mini",
    "ollama": "gemma4",
}


# ── Model factory ───────────────────────────────────────────────────────────


def _build_model(provider: str, model_name: str, host: str) -> Model:
    if provider == "google":
        return GoogleModel(
            model_name=model_name or PROVIDER_DEFAULTS["google"],
            provider=GoogleProvider(),
        )
    elif provider == "openai":
        return OpenAIChatModel(
            model_name=model_name or PROVIDER_DEFAULTS["openai"],
            provider=OpenAIProvider(),
        )
    elif provider == "ollama":
        base = host.rstrip("/")
        if not base.endswith("/v1"):
            base += "/v1"
        return OpenAIChatModel(
            model_name=model_name or PROVIDER_DEFAULTS["ollama"],
            provider=OllamaProvider(base_url=base),
        )
    else:
        raise ValueError(f"unsupported provider: {provider}")


def create_model(
    provider: str,
    model_name: str,
    host: str,
    fallback: str = "",
) -> Model:
    primary = _build_model(provider, model_name, host)
    if not fallback:
        return primary
    secondary = _build_model(fallback, "", host)
    return FallbackModel(primary, secondary)


# ── Artifact generation (for corpus) ────────────────────────────────────────


def clean_artifact(text: str) -> str:
    text = text.strip()
    if text.startswith("```"):
        nl = text.find("\n")
        if nl != -1:
            text = text[nl + 1 :]
    if text.endswith("```"):
        text = text[:-3].rstrip()
    text = re.sub(r"<think>.*?</think>", "", text, flags=re.DOTALL).strip()
    return text


def generate_artifact(model: Model, prompt: str) -> str:
    """Generate a single artifact. Returns cleaned content."""
    agent: Agent[None, str] = Agent(
        model,
        system_prompt="You are a code generator. Output only raw code/content. No markdown fences, no explanation.",
    )
    result = agent.run_sync(prompt)
    return clean_artifact(result.output)
