> Permissions Proposal Draft 1.

# Rules

1. A negative permission overrides a positive one at the same or higher level.
1. Absence of a positive permission is equal to a blanket negative at that level.

## Implications

* An object is by default not shared with anything.
* Permissions on an object override those of the class.

# Hierarchies

## Scope

* NONE - Only me, "not shared"
* SOME - Specific networks
* NETS - All my networks
* PUBLIC - Public

## Types

* OBJECT - Per-object
* CLASS - Per-class

# Actors

## Alice

Is in three networks: Stonebroti, Morfi, Boundgrave.

Has three skills:

* Alchemy is shared to public (OBJECT-PUBLIC)
* Acrobatics is shared only to Boundgrave (OBJECT-SOME)
* Archery is not shared at all (OBJECT-NONE)

## Bob

Is in two networks: Boundgrave, Terregonje

Has three skills:

* Brainwashing is not shared at all (OBJECT-NONE)
* Boating is shared to all his networks (OBJECT-NETS)
* Birdwatching is shared to public (OBJECT-PUBLIC)

## Chip

Is in two networks: Mextunmo, Terregonje

Has three skills, shared as a class to all his networks, but blocking Cooking from Mextunmo.

* Alchemy (CLASS-NETS)
* Cooking (CLASS-NETS, OBJECT-SOME !Mextunmo)
* Criminology (CLASS-NETS)

## Diana

Is in two networks: Mextunmo, Terregonje

Has three skills:

* Dancing, shared to all her networks except Terregonje (OBJECT-NETS, !Terregonje)
* Diplomacy, shared to public (OBJECT-PUBLIC)
* Disguise, not shared (OBJECT-NONE)

## Frank

Is in two networks: Toompark, Percalcombe

Has three skills which he shares as a class with all his networks:

* Falconry (CLASS-NETS)
* Forgery (CLASS-NETS)
* Forensics (CLASS-NETS)

# Who can see what?

## Alice

* can see Bob's skill in Boating (OBJECT-NETS, Boundgrave) and Birdwatching (OBJECT-PUBLIC)
* can see none of Chip's skills (OBJECT-NETS, no shared network)
* can see Diana's skill in Diplomacy (OBJECT-PUBLIC)
* can see none of Frank's skills (CLASS-NETS, no shared network)

## Bob

* can see Alice's skill in Alchemy (OBJECT-PUBLIC) and Acrobatics (OBJECT-NETS, Boundgrave)
* can see all of Chip's skills (CLASS-NETS, Terregonje)
* can see Diana's skill in Diplomacy (OBJECT-PUBLIC) but not Dancing (OBJECT-NETS, !Terregonje)
* can see none of Frank's skills (CLASS-NETS, no shared network)

## Chip

* can see Alice's skill in Alchemy (OBJECT-PUBLIC)
* can see Bob's skill in Boating (OBJECT-NETS, Terregonje) and Birdwatching
* can see Diana's skill in Diplomacy (OBJECT-PUBLIC) and Dancing (OBJECT-NETS, Mextunmo)
* can see none of Frank's skills (CLASS-NETS, no shared network)

## Diana

* can see Alice's skill in Alchemy (OBJECT-PUBLIC)
* can see Bob's skill in Boating (OBJECT-NETS, Terregonje) and Birdwatching
* can see all of Chip's skills (OBJECT-NETS, Mextunmo or Terregonje)
* can see none of Frank's skills (CLASS-NETS, no shared network)

## Frank

* can see Alice's skill in Alchemy (OBJECT-PUBLIC)
* can see Bob's skill in Birdwatching (OBJECT-PUBLIC)
* can see none of Chip's skills (CLASS-NETS)
* can see Diana's skill in Diplomacy (OBJECT-PUBLIC)

## Summary

Left column is viewer, top row is viewee; ie. Alice can see Diana's Diplomacy.

| | Alice | Bob | Chip | Diana | Frank |
|-----|-----|-----|-----|-----|-----|
| Alice | --- | Boating, Birdwatching | none | Diplomacy | none |
| Bob | Alchemy, Acrobatics | --- | all | Diplomacy | none |
| Chip | Alchemy | Boating, Birdwatching | --- | Diplomacy, Dancing | none |
| Diana | Alchemy | Boating, Birdwatching | all | --- | none |
| Frank | Alchemy | Birdwatching | none | Diplomacy | --- |

# Implementation

A suggested implementation which provides the functionality we need whilst also allowing future changes to be made without extensive re-work.

```sql
CREATE TABLE perms AS (
    user_id     BIGINT,
    object_type VARCHAR(16),
    object_id   BIGINT,
    class       BOOLEAN,
    public      BOOLEAN,
    hidden      BOOLEAN,
    perms       JSONB
)
```

The JSON field would have this structure.

```json
{
  "networks": {
    "shared_to": [1,2,3,4], "hidden_from": [5,6], "all": true
  }
}
```

That can be unmarshalled into a Go structure.

```go
type Permission struct {
	SharedTo   []int64
	HiddenFrom []int64
	All        bool
}

type ObjectPermission struct {
	UserID     int64
	ObjectType string
	ObjectID   int64
	Class      bool
	Public     bool
	Hidden     bool
    Masks      map[string]Permission
}
```

## Pseudcode

> Note that there can be two permissions for an object - one at object level (ie. on a specific skill) and the other at class level (ie.  for all my skills).

In which VIEWER is looking at VIEWEE's skill S.

```
Fetch Object and/or Class permissions for S for VIEWEE into L
For each network N from VIEWER, set V[N] = false
For each permission P in list L ordered by Class, Object:
  If P.HIDDEN:
    set all in V[] = false
  If P.NETWORKS.ALL:
    set P.NETWORKS.SHARED_TO = all networks from VIEWEE
  For all networks N in P.NETWORKS.SHARED_TO:
    If V[N] exists:
      set V[N] = true
  For all networks N in V[]:
    set V[N] = V[N] || P.PUBLIC
  For all networks N in P.NETWORKS.HIDDEN_FROM:
    set V[N] to false
Visible = Logical-Or of V[]
```

### Worked example: Bob looking at Diana's profile

Diana's database entries would look something like this - using names instead of object ids
and abbreviating the JSON with the assumption that missing is empty or `false`.

| User | Type | Object | Class? | Perms |
|-----|-----|-----|-----|----|
| Diana | skill | Dancing | false | `{"networks":{"hidden_from":["Terregonje"]},"ALL":true}` |
| Diana | skill | Diplomacy | false | `{"public":true}` |
| Diana | skill | Disguise | false | `{"hidden": true}` |

Bob and Diana have the Terregonje network in common.

#### Dancing

1. V[Terregonje] = false, V[Boundgrave] = false
1. Only one row to consider at Object level
1. P.HIDDEN = false, skipped
1. P.NETWORKS.ALL = true, P.NETWORKS.SHARED_TO = [Terregonje, Mextunmo]
1. Terregonje in P.NETWORKS.SHARED_TO, V[Terregonje] = true
1. Mextunmo in P.NETWORKS.SHARED_TO, V[Mextunmo] missing, skipped
1. P.PUBLIC = false, V[] is unchanged
1. Terregonje in P.NETWORKS.HIDDEN_FROM, V[Terregonje] = false
1. Visible = V[Terregonje] || V[Boundgrave] = false

Bob cannot see Diana's skill in Dancing because whilst it is shared to all her networks,
it is blocked from Terregonje which is the only network Bob has in common.

#### Diplomacy

1. V[Terregonje] = false, V[Boundgrave] = false
1. Only one row to consider at Object level
1. P.HIDDEN = false, skipped
1. P.NETWORKS.ALL = false, skipped
1. Nothing in P.NETWORKS.SHARED_TO, skipped
1. P.PUBLIC = true, V[Terregonje] = true, V[Boundgrave] = true
1. Nothing in P.NETWORKS.HIDDEN_FROM, skipped
1. Visible = V[Terregonje] || V[Boundgrave] = true

Bob can see Diana's skill in Diplomacy because it is public.

#### Disguise

1. V[Terregonje] = false, V[Boundgrave] = false
1. Only one row to consider at Object level
1. P.HIDDEN = true, V[Terregonje] = false, V[Boundgrave] = false
1. P.NETWORKS.ALL = false, skipped
1. Nothing in P.NETWORKS.SHARED_TO, skipped
1. P.PUBLIC = false, V[Terregonje] unchanged
1. Nothing in P.NETWORKS.HIDDEN_FROM, skipped
1. Visible = V[Terregonje] || V[Boundgrave] = false

Bob cannot see Diana's skill in Disguise because it is hidden from everyone.

### Worked example: Chip looking at Diana's Dancing

1. V[Mextunmo] = false, V[Terregonje] = false
1. Only one row to consider at Object level
1. P.HIDDEN = false, skipped
1. P.NETWORKS.ALL = true, P.NETWORKS.SHARED_TO = [Mextunmo, Terregonje]
1. Terregonje in P.NETWORKS.SHARED_TO, V[Terregonje] = true
1. Mextunmo in P.NETWORKS.SHARED_TO, V[Mextunmo] = true
1. P.PUBLIC = false, V[] is unchanged
1. Terregonje in P.NETWORKS.HIDDEN_FROM, V[Terregonje] = false
1. Visible = V[Mextunmo] || V[Terregonje] = true

Chip can see Diana's skill in Dancing because whilst it is blocked from Terregonje, he also has the Mextunmo network in common.

### Worked example: Frank looking at Diana's profile

Frank and Diana have no networks in common.

#### Dancing

1. V[Toompark] = false, V[Percalcombe] = false
1. Only one row to consider at Object level
1. P.HIDDEN = false, skipped
1. P.NETWORKS.ALL = true, P.NETWORKS.SHARED_TO = [Mextunmo, Terregonje]
1. Terregonje in P.NETWORKS.SHARED_TO, V[Terregonje] missing, skipped
1. Mextunmo in P.NETWORKS.SHARED_TO, V[Mextunmo] missing, skipped
1. P.PUBLIC = false, V[] is unchanged
1. No common networks in P.NETWORKS.HIDDEN_FROM, skipped
1. Visible = V[Toompark] || V[Percalcombe] = false

Frank cannot see Diana's skill in Dancing because they have no networks in common.

#### Diplomacy

1. V[Toompark] = false, V[Percalcombe] = false
1. Only one row to consider at Object level
1. P.HIDDEN = false, skipped
1. P.NETWORKS.ALL = false, skipped
1. Nothing in P.NETWORKS.SHARED_TO, skipped
1. P.PUBLIC = true, V[Toompark] = true, V[Percalcombe] = true
1. Nothing in P.NETWORKS.HIDDEN_FROM, skipped
1. Visible = V[Toompark] || V[Percalcombe] = true

Frank can see Diana's skill in Diplomacy because it is public.

#### Disguise

1. V[Toompark] = false, V[Percalcombe] = false
1. Only one row to consider at Object level
1. P.HIDDEN = true, V[...] = false
1. P.NETWORKS.ALL = false, skipped
1. Nothing in P.NETWORKS.SHARED_TO, skipped
1. P.PUBLIC = false, V[] unchanged
1. Nothing in P.NETWORKS.HIDDEN_FROM, skipped
1. Visible = V[Toompark] || V[Percalcombe] = false

Frank cannot see Diana's skill in Disguise because it is hidden from everyone.

### Worked example: Diana looking at Chip's profile

Chip's database entries would look something like this - using names instead of object ids and abbreviating the JSON with the assumption that missing is empty or `false`.

| User | Type | Object | Class? | Perms |
|-----|-----|-----|-----|----|
| Chip | skill | * | true | `{"ALL":true}` |
| Chip | skill | Cooking | false | `{"networks":{"hidden_from":"Mextunmo"}}` |

Chip and Diana have both the Mextunmo and Terregonje networks in common.

#### Alchemy, Criminology

2. V[Mextunmo] = false, V[Terregonje] = false
3. Only one row to consider at Class level
4. P.HIDDEN = false, skipped
5. P.NETWORKS.ALL = true, P.NETWORKS.SHARED_TO = [Mextunmo, Terregonje]
6. Terregonje in P.NETWORKS.SHARED_TO, V[Terregonje] = true
7. Mextunmo in P.NETWORKS.SHARED_TO, V[Mextunmo] = true
7. P.PUBLIC = false, V[] is unchanged
8. Nothing in P.NETWORKS.HIDDEN_FROM, skipped
9. Visible = V[Mextunmo] || V[Terregonje] = true

Diana can see both Alchemy and Criminology in Chip's skills because they are shared as a class to his networks and she is in two common networks.

#### Cooking

2. V[Mextunmo] = false, V[Terregonje] = false
3. Two rows to consider at both Class (C) and Object (O) level
4. C.P.HIDDEN = false, skipped
5. C.P.NETWORKS.ALL = true, C.P.NETWORKS.SHARED_TO = [Mextunmo, Terregonje]
6. Terregonje in C.P.NETWORKS.SHARED_TO, V[Terregonje] = true
7. Mextunmo in C.P.NETWORKS.SHARED_TO, V[Mextunmo] = true
7. C.P.PUBLIC = false, V[] is unchanged
8. Nothing in C.P.NETWORKS.HIDDEN_FROM, skipped
4. O.P.HIDDEN = false, skipped
5. O.P.NETWORKS.ALL = false, skipped
6. O.P.NETWORKS.SHARED_TO empty, skipped
7. O.P.PUBLIC = false, V[] is unchanged
8. Mextunmo in O.P.NETWORKS.HIDDEN_FROM, V[Mextunmo] = false
9. Visible = V[Mextunmo] || V[Terregonje] = true

Diana can also see Chip's skill in Cooking because whilst the Mextunmo network is blocked, she gets visibility through Terregonje.

# Questions

* What happens if Alice and Bob have one network N in common and Alice has made a skill "Public but not N"?
