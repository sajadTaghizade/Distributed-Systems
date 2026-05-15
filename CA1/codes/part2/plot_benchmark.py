from __future__ import annotations

import re
from io import StringIO
from pathlib import Path

import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
import seaborn as sns


sns.set_theme(style="whitegrid", context="paper", font_scale=1.15)
plt.rcParams.update(
    {
        "figure.dpi": 300,
        "savefig.dpi": 300,
        "font.family": "DejaVu Sans",
        "axes.titleweight": "bold",
    }
)


def extract_markdown_tables(text: str) -> list[str]:
    blocks: list[str] = []
    current: list[str] = []
    for line in text.splitlines():
        if "|" in line:
            current.append(line.rstrip())
        else:
            if current:
                blocks.append("\n".join(current))
                current = []
    if current:
        blocks.append("\n".join(current))
    return blocks


def parse_markdown_table(block: str) -> pd.DataFrame | None:
    lines = [line.strip() for line in block.splitlines() if line.strip()]
    if len(lines) < 2:
        return None
    if "-" not in lines[1].replace(" ", ""):
        return None

    rows: list[list[str]] = []
    for idx, line in enumerate(lines):
        if idx == 1:
            continue
        cells = [cell.strip() for cell in line.strip("|").split("|")]
        if len(cells) >= 2:
            rows.append(cells)

    if len(rows) < 2:
        return None

    width = max(len(row) for row in rows)
    rows = [row + [""] * (width - len(row)) for row in rows]
    return pd.DataFrame(rows[1:], columns=rows[0])


def load_data(markdown_path: Path) -> pd.DataFrame:
    text = markdown_path.read_text(encoding="utf-8")
    frames: list[pd.DataFrame] = []
    for block in extract_markdown_tables(text):
        df = parse_markdown_table(block)
        if df is not None and not df.empty:
            frames.append(df)
    if not frames:
        return pd.read_csv(StringIO(text))
    return pd.concat(frames, ignore_index=True)


def normalize_columns(df: pd.DataFrame) -> pd.DataFrame:
    def normalize_name(name: str) -> str:
        return re.sub(r"[^a-z0-9]+", "", str(name).lower())

    rename_map: dict[str, str] = {}
    for col in df.columns:
        key = normalize_name(col)
        if any(token in key for token in ("gomaxprocs", "gmp", "procs")):
            rename_map[col] = "GOMAXPROCS"
        elif any(token in key for token in ("goroutine", "concurrency", "threads", "workers")):
            rename_map[col] = "Goroutines"
        elif "workload" in key or "type" in key:
            rename_map[col] = "Workload"
        elif "throughput" in key or "opssec" in key or "opspersecond" in key:
            rename_map[col] = "Throughput_ops_s"
        elif "stddev" in key or "standarddeviation" in key or "stdev" in key or key == "sd":
            rename_map[col] = "StdDev"
        elif "latency" in key and "max" in key:
            rename_map[col] = "Max_Latency"
        elif "latency" in key and ("avg" in key or "average" in key):
            rename_map[col] = "Average_Latency"
        elif "latency" in key:
            rename_map[col] = "Average_Latency"

    df = df.rename(columns=rename_map).copy()
    required = {"GOMAXPROCS", "Goroutines", "Workload"}
    missing = required - set(df.columns)
    if missing:
        raise ValueError(f"Missing required columns after normalization: {sorted(missing)}")
    return df


def parse_number(value) -> float:
    if pd.isna(value):
        return np.nan
    if isinstance(value, (int, float, np.integer, np.floating)):
        return float(value)
    text = str(value).strip().replace(",", "")
    if not text:
        return np.nan
    match = re.search(r"[-+]?\d*\.?\d+(?:[eE][-+]?\d+)?", text)
    return float(match.group()) if match else np.nan


def parse_time_ms(value) -> float:
    if pd.isna(value):
        return np.nan
    if isinstance(value, (int, float, np.integer, np.floating)):
        return float(value)
    text = str(value).strip().replace(",", "")
    if not text:
        return np.nan
    number = parse_number(text)
    lower = text.lower().replace("μ", "µ")
    if "µs" in lower or "us" in lower:
        return number / 1000.0
    if "ns" in lower:
        return number / 1_000_000.0
    if "ms" in lower:
        return number
    if re.search(r"\bs\b", lower) and "ops/s" not in lower and "sec" not in lower:
        return number * 1000.0
    return number


def coerce_columns(df: pd.DataFrame) -> pd.DataFrame:
    df = df.copy()
    df["GOMAXPROCS"] = df["GOMAXPROCS"].map(parse_number).round().astype("Int64")
    df["Goroutines"] = df["Goroutines"].map(parse_number).round().astype("Int64")
    df["Workload"] = (
        df["Workload"]
        .astype(str)
        .str.replace(r"[*_`]+", "", regex=True)
        .str.strip()
    )
    if "Throughput_ops_s" in df.columns:
        df["Throughput_ops_s"] = df["Throughput_ops_s"].map(parse_number)
    if "Average_Latency" in df.columns:
        df["Average_Latency_ms"] = df["Average_Latency"].map(parse_time_ms)
    if "StdDev" in df.columns:
        df["StdDev_ms"] = df["StdDev"].map(parse_time_ms)
    return df


def aggregate_data(df: pd.DataFrame) -> pd.DataFrame:
    metrics = [c for c in ["Throughput_ops_s", "Average_Latency_ms", "StdDev_ms"] if c in df.columns]
    return (
        df.dropna(subset=["GOMAXPROCS", "Goroutines", "Workload"])
        .groupby(["GOMAXPROCS", "Workload", "Goroutines"], as_index=False)
        .agg({metric: "mean" for metric in metrics})
        .sort_values(["GOMAXPROCS", "Workload", "Goroutines"])
    )


def save_plot_1(df: pd.DataFrame, out_dir: Path) -> None:
    plot_df = df[df["GOMAXPROCS"].isin([1, 12]) & df["Workload"].isin(["CPU-bound", "Mixed"])].copy()
    g_order = sorted(plot_df["Goroutines"].dropna().unique().tolist())
    palette = {"CPU-bound": "#1f77b4", "Mixed": "#d62728"}
    markers = {"CPU-bound": "o", "Mixed": "s"}

    fig, axes = plt.subplots(1, 2, figsize=(14, 5.5), sharey=True, constrained_layout=True)
    for ax, procs in zip(axes, [1, 12]):
        sub = plot_df[plot_df["GOMAXPROCS"] == procs].copy()
        sub["Goroutines"] = pd.Categorical(sub["Goroutines"], categories=g_order, ordered=True)
        sub = sub.sort_values("Goroutines")
        for workload in ["CPU-bound", "Mixed"]:
            wsub = sub[sub["Workload"] == workload]
            if wsub.empty:
                continue
            ax.plot(
                wsub["Goroutines"].astype(int),
                wsub["Throughput_ops_s"],
                label=workload,
                color=palette[workload],
                marker=markers[workload],
                linewidth=2.2,
                markersize=6,
            )
        ax.set_xscale("log", base=2)
        ax.set_xticks(g_order)
        ax.get_xaxis().set_major_formatter(plt.ScalarFormatter())
        ax.set_xlabel("Number of Goroutines")
        ax.set_title(f"GOMAXPROCS = {procs}")
        ax.grid(True, axis="y", alpha=0.25)
        ax.legend(frameon=True, title="Workload")

    axes[0].set_ylabel("Throughput (ops/sec)")
    fig.suptitle("Scalability & Throughput", fontsize=15, fontweight="bold")
    fig.savefig(out_dir / "plot1_scalability_throughput.png", bbox_inches="tight")
    plt.close(fig)


def save_plot_2(df: pd.DataFrame, out_dir: Path) -> None:
    plot_df = df[df["Workload"].astype(str).str.strip() == "Mixed"].copy()
    g_order = sorted(plot_df["Goroutines"].dropna().unique().tolist())
    hue_order = [g for g in [1, 2, 12] if g in plot_df["GOMAXPROCS"].dropna().unique().tolist()]

    fig, ax = plt.subplots(figsize=(13, 5.8), constrained_layout=True)
    sns.barplot(
        data=plot_df,
        x="Goroutines",
        y="StdDev_ms",
        hue="GOMAXPROCS",
        order=g_order,
        hue_order=hue_order,
        palette="crest",
        errorbar=None,
        ax=ax,
    )
    ax.set_title("The Cost of Context Switching — Variance Spike")
    ax.set_xlabel("Number of Goroutines")
    ax.set_ylabel("Standard Deviation (ms)")
    ax.legend(title="GOMAXPROCS", frameon=True)
    ax.grid(True, axis="y", alpha=0.25)
    fig.savefig(out_dir / "plot2_context_switching_variance_spike.png", bbox_inches="tight")
    plt.close(fig)


def save_plot_3(df: pd.DataFrame, out_dir: Path) -> None:
    plot_df = df[df["GOMAXPROCS"] == 12].copy()
    plot_df = plot_df[plot_df["Workload"].isin(["CPU-bound", "Mixed"])]
    g_order = sorted(plot_df["Goroutines"].dropna().unique().tolist())

    fig, ax = plt.subplots(figsize=(12, 5.5), constrained_layout=True)
    sns.lineplot(
        data=plot_df,
        x="Goroutines",
        y="Average_Latency_ms",
        hue="Workload",
        style="Workload",
        markers=True,
        dashes=False,
        palette={"CPU-bound": "#1f77b4", "Mixed": "#d62728"},
        hue_order=["CPU-bound", "Mixed"],
        style_order=["CPU-bound", "Mixed"],
        ax=ax,
    )
    ax.set_xscale("log", base=2)
    ax.set_xticks(g_order)
    ax.get_xaxis().set_major_formatter(plt.ScalarFormatter())
    ax.set_title("Latency vs. Concurrency (GOMAXPROCS = 12)")
    ax.set_xlabel("Number of Goroutines")
    ax.set_ylabel("Average Latency (ms)")
    ax.legend(title="Workload", frameon=True)
    ax.grid(True, axis="y", alpha=0.25)
    fig.savefig(out_dir / "plot3_latency_vs_concurrency.png", bbox_inches="tight")
    plt.close(fig)


def main() -> None:
    base_dir = Path(__file__).resolve().parent
    markdown_path = base_dir / "2.md"
    output_dir = base_dir
    raw_df = load_data(markdown_path)
    df = aggregate_data(coerce_columns(normalize_columns(raw_df)))
    save_plot_1(df, output_dir)
    save_plot_2(df, output_dir)
    save_plot_3(df, output_dir)
    print("Saved plots to:", output_dir)


if __name__ == "__main__":
    main()
