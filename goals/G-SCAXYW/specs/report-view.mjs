// The view of report.tla for the statechart: where the tool and the grill
// are, and what the select did, in the owner's words.

export const title = "hyp G-SCAXYW: the tool's report and the grill's tries";

const work = (v) => {
  switch (v.phase) {
    case "start": return "Model written";
    case "configs": return "Variant configs written";
    case "walks": return v.fixed ? "Model fixed, walks matched" : "Walks matched";
    case "cells": return "Cells predicted";
    case "tried": return "Variants tried";
    case "report": return v.diff.length ? "Report: a try differs" : "Report: every try agrees";
    case "asked": return "Owner asked";
  }
  return v.phase;
};

const select = (v) => (v.sel === "-" ? "Not selected" : v.sel === "recorded" ? "Select recorded" : "Select refused");

export const state = (v) => ({ "Tool and grill": work(v), Select: select(v) });

export const group = (v) => (v.phase === "asked" ? "done" : ["start", "configs", "walks", "cells"].includes(v.phase) ? "tool" : "grill");

export const colors = { tool: "teal", grill: "orange", done: "purple" };

export const step = (action) => {
  switch (action) {
    case "ToolWritesConfigs": return "the tool runs each rule alone under each variant";
    case "ToolMatchesWalks": return "the tool matches each floor to a walk";
    case "ToolPredicts": return "the tool predicts each example under each variant";
    case "GrillTries": return "the grill tries the variants with hyp try";
    case "ToolReports": return "the tool writes the report";
    case "GrillFixes": return "the grill fixes the model";
    case "GrillAsks": return "the grill asks the owner";
    case "Selects": return "a select reads the spec folder";
    case "Done": return null;
  }
  return action;
};
