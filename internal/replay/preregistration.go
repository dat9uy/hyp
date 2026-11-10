package replay

import "github.com/dat9uy/hyp/internal/event"

// An Experiment is a pre-registered Result (ADR 0018, spec §4
// pre-registration). Before its run it fixes its targets, aims, recipe,
// rubric and records, and none of them changes after the add. This file is
// the one place that reads those facts as a whole: the targets a Result must
// give a verdict on, the question an Experiment belongs to, and its shape.

// Shapes of an Experiment, derived from its parent kind, targets and rubric,
// never stored.
const (
	ShapeRunAndRule = "run and rule"
	ShapeReading    = "rule by reading"
	ShapeReassess   = "reassess from kept records"
	ShapeMeasure    = "measure only"
)

// targetsOf lists the targets of Experiment n: its parent Hypothesis, then
// each Hypothesis in its targets field. An Experiment under a DesignQuestion
// has none. Every Result of n gives a verdict on each target and on nothing
// else (spec §5 rule 6).
func targetsOf(nodes map[string]*Node, n *Node) []string {
	if p := nodes[n.Parent]; p == nil || p.Kind != event.KindHypothesis {
		return nil
	}
	return append([]string{n.Parent}, n.Targets...)
}

// questionOf is the DesignQuestion Experiment n belongs to: its parent when
// it sits under a question, otherwise its parent Hypothesis's question.
func questionOf(nodes map[string]*Node, n *Node) string {
	p := nodes[n.Parent]
	if p == nil {
		return ""
	}
	if p.Kind == event.KindDesignQuestion {
		return p.ID
	}
	return p.Parent
}

// shapeOf derives Experiment n's shape (spec §4 pre-registration): under a
// DesignQuestion it measures only; with a rubric bundle it runs and rules
// when it pins a recipe and reassesses when it does not; with a prose rubric
// it reassesses when it has no recipe and pins records, and rules by reading
// otherwise. An Experiment from before ADR 0018 with neither rubric nor
// records reads as run and rule when it pins a recipe.
func shapeOf(nodes map[string]*Node, n *Node) string {
	if p := nodes[n.Parent]; p != nil && p.Kind == event.KindDesignQuestion {
		return ShapeMeasure
	}
	noRecipe := n.Recipe == event.RecipeNone
	switch {
	case n.Rubric == event.RubricReading:
		if noRecipe && len(n.Records) > 0 {
			return ShapeReassess
		}
		return ShapeReading
	case n.Rubric != "" && noRecipe:
		return ShapeReassess
	case noRecipe:
		return ShapeReading
	}
	return ShapeRunAndRule
}

// TargetsOf lists Experiment exp's targets (targetsOf).
func (p Projection) TargetsOf(exp string) []string { return targetsOf(p.Nodes, p.Nodes[exp]) }

// ShapeOf names Experiment exp's shape (shapeOf).
func (p Projection) ShapeOf(exp string) string { return shapeOf(p.Nodes, p.Nodes[exp]) }

// QuestionOf is the DesignQuestion Experiment exp belongs to (questionOf).
func (p Projection) QuestionOf(exp string) string { return questionOf(p.Nodes, p.Nodes[exp]) }
