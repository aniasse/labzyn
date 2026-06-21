package twentyapi

import (
	"time"

	"github.com/aniasse/labzyn/server/internal/metadata"
)

// Twenty's frontend assumes certain standard objects always exist (workspaceMember
// is referenced everywhere for actors/avatars). We synthesize them so the app
// renders; their records resolve empty for now.

const workspaceMemberObjectID = "20202020-1c25-4d02-bf25-6aeccf7ea419"

func standardObjects() []objectWithFields {
	now := time.Now()
	wmFields := []metadata.Field{
		stdField(workspaceMemberObjectID, "name", "Name", metadata.FieldType("FULL_NAME"), now),
		stdField(workspaceMemberObjectID, "colorScheme", "Color Scheme", metadata.FieldText, now),
		stdField(workspaceMemberObjectID, "avatarUrl", "Avatar Url", metadata.FieldText, now),
		stdField(workspaceMemberObjectID, "locale", "Language", metadata.FieldText, now),
		stdField(workspaceMemberObjectID, "userEmail", "User Email", metadata.FieldText, now),
		stdField(workspaceMemberObjectID, "timeZone", "Time Zone", metadata.FieldText, now),
		stdField(workspaceMemberObjectID, "dateFormat", "Date Format", metadata.FieldText, now),
		stdField(workspaceMemberObjectID, "timeFormat", "Time Format", metadata.FieldText, now),
	}
	wm := metadata.Object{
		ID:            workspaceMemberObjectID,
		NameSingular:  "workspaceMember",
		NamePlural:    "workspaceMembers",
		LabelSingular: "Workspace Member",
		LabelPlural:   "Workspace Members",
		Labels:        []byte("{}"),
		Icon:          "IconUserCircle",
		IsSystem:      true,
		IsActive:      true,
		Storage:       "jsonb",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	return []objectWithFields{{Object: wm, Fields: wmFields}}
}

type objectWithFields struct {
	Object metadata.Object
	Fields []metadata.Field
}

func stdField(objectID, name, label string, t metadata.FieldType, now time.Time) metadata.Field {
	return metadata.Field{
		ID:         string(synthFieldID(objectID, name)),
		ObjectID:   objectID,
		Name:       name,
		Type:       t,
		Label:      label,
		IsSystem:   true,
		IsActive:   true,
		IsNullable: true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
