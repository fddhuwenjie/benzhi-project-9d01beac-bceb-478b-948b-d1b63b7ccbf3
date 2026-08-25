package domain

import (
	"reflect"
	"time"
)

type BlockFieldChange struct {
	Field  string `json:"field"`
	Before any    `json:"before,omitempty"`
	After  any    `json:"after,omitempty"`
}

type BlockDifference struct {
	BlockID   string             `json:"block_id"`
	FromIndex int                `json:"from_index,omitempty"`
	ToIndex   int                `json:"to_index,omitempty"`
	Changes   []BlockFieldChange `json:"changes,omitempty"`
}

type ContentDifference struct {
	FromRevision int64             `json:"from_revision"`
	ToRevision   int64             `json:"to_revision"`
	Added        []BlockDifference `json:"added"`
	Deleted      []BlockDifference `json:"deleted"`
	Moved        []BlockDifference `json:"moved"`
	Changed      []BlockDifference `json:"changed"`
}

func CompareContent(from, to DocumentContent) ContentDifference {
	result := ContentDifference{FromRevision: from.Revision, ToRevision: to.Revision, Added: []BlockDifference{}, Deleted: []BlockDifference{}, Moved: []BlockDifference{}, Changed: []BlockDifference{}}
	fromByID := make(map[string]struct {
		block ContentBlock
		index int
	}, len(from.Blocks))
	toByID := make(map[string]struct {
		block ContentBlock
		index int
	}, len(to.Blocks))
	for index, block := range from.Blocks {
		fromByID[block.ID] = struct {
			block ContentBlock
			index int
		}{block, index}
	}
	for index, block := range to.Blocks {
		toByID[block.ID] = struct {
			block ContentBlock
			index int
		}{block, index}
	}
	for _, block := range from.Blocks {
		before := fromByID[block.ID]
		after, exists := toByID[block.ID]
		if !exists {
			result.Deleted = append(result.Deleted, BlockDifference{BlockID: block.ID, FromIndex: before.index})
			continue
		}
		if before.index != after.index {
			result.Moved = append(result.Moved, BlockDifference{BlockID: block.ID, FromIndex: before.index, ToIndex: after.index})
		}
		changes := compareBlock(before.block, after.block)
		if len(changes) > 0 {
			result.Changed = append(result.Changed, BlockDifference{BlockID: block.ID, FromIndex: before.index, ToIndex: after.index, Changes: changes})
		}
	}
	for _, block := range to.Blocks {
		if _, exists := fromByID[block.ID]; !exists {
			result.Added = append(result.Added, BlockDifference{BlockID: block.ID, ToIndex: toByID[block.ID].index})
		}
	}
	return result
}

func compareBlock(before, after ContentBlock) []BlockFieldChange {
	var changes []BlockFieldChange
	add := func(field string, old, current any) {
		if !reflect.DeepEqual(old, current) {
			changes = append(changes, BlockFieldChange{Field: field, Before: old, After: current})
		}
	}
	add("type", before.Type, after.Type)
	add("heading_level", before.HeadingLevel, after.HeadingLevel)
	add("text", before.Text, after.Text)
	add("link_target", before.LinkTarget, after.LinkTarget)
	add("alt_text", before.AltText, after.AltText)
	add("table_headers", before.TableHeaders, after.TableHeaders)
	return changes
}

type AuditDifference struct {
	ContentRevision int64    `json:"content_revision"`
	Resolved        []string `json:"resolved"`
	Persistent      []string `json:"persistent"`
	Added           []string `json:"added"`
}

func (c *AcceptanceCase) Reaudit(issues []AuditIssue, version string, now time.Time) (AuditDifference, error) {
	if c.Status != StatusRemediating || c.ContentRevision <= c.AuditContentRevision || c.ContentRevision == 0 {
		return AuditDifference{}, ErrNotReady
	}
	previous := make(map[string]AuditIssue, len(c.Issues))
	for _, issue := range c.Issues {
		previous[issue.RuleCode+"\x00"+issue.BlockID] = issue
	}
	current := make(map[string]bool, len(issues))
	difference := AuditDifference{ContentRevision: c.ContentRevision, Resolved: []string{}, Persistent: []string{}, Added: []string{}}
	for i := range issues {
		key := issues[i].RuleCode + "\x00" + issues[i].BlockID
		current[key] = true
		if old, exists := previous[key]; exists {
			difference.Persistent = append(difference.Persistent, old.ID)
		} else {
			difference.Added = append(difference.Added, issues[i].ID)
		}
	}
	for key, issue := range previous {
		if !current[key] {
			difference.Resolved = append(difference.Resolved, issue.ID)
		}
	}
	c.IssueHistory = append(c.IssueHistory, cloneIssues(c.Issues)...)
	c.Issues = cloneIssues(issues)
	c.RuleVersion = version
	c.AuditContentRevision = c.ContentRevision
	c.LastAuditDifference = &difference
	c.Touch(now)
	c.RefreshReviewStatus(now)
	return difference, nil
}

func cloneIssues(issues []AuditIssue) []AuditIssue {
	out := make([]AuditIssue, len(issues))
	copy(out, issues)
	for i := range out {
		if issues[i].Evidence != nil {
			evidence := *issues[i].Evidence
			out[i].Evidence = &evidence
		}
	}
	return out
}
