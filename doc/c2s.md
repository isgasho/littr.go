# littr.me

We have two types of endpoints that we can query on FedBOX:

## Object collections:

The object collections in the ActivitypPub spec are: following, followers, liked.
Additionally FedBOX supports `/actors` and `/objects` generic objects end-points.

Supported filters:

  * PublishedDate
  * Type 
  * To
  * BCC
  * Audience
  * Generator[1]
  * Url[2]

[1] Generator is not yet implemented in littr.go and it should contain the IRI of the C2S application that generated the Activity  
[2] Url is not yet implemented in littr.go and it should probably contain a URL from the generator application  

## Activity collections:

The activity collections in the ActivitypPub spec are: outbox, inbox, likes, shares, replies.
Additionally FedBOX supports `/activities` generic activities end-point.

In order to get the full representation of the items, after loading one of these collections, their Object properties need to be dereferenced and loaded again.

Besides the filters applicable to Object collections we have also:

  * Actor IRI
  * Object IRI
  * Target IRI

# The filtering

Filtering collections is done using query parameters corresponding to the lowercased value of the property's name it matches against.

All filters can be appended multiple times in the URL. 

The matching logic is:

* Multiple values of same filter are matched by doing an union on the resulting sets.
* Different filters match by doing an intersection on the resulting sets of each filter.

We currently need to add the possibility of prepending operators to the values, so we can support negative matches, or some other types of filtering.

## Querying FedBOX

### Loading main page items

Loads FedBOX's service inbox `/inbox`

### Loading federated tab items

Same as main page items, but the filtering should have the base IRI different than fedbox's host.[1]


### Loading followed items

Loads the logged account's inbox. `/actors/{uuid}/inbox`

### Load discussions

Loads the `/objects` end-point with a filter for Url being empty and Context being empty.

### Load items from a particular domain

Loads the `/objects` end-point with a filter on Url to match the required domain.

### Load a user's items

Loads the outbox of the actor corresponding to the user: `/actors/{uuid}/outbox`

### Load a user's votes

Loads the liked end-point of the actor corresponding to the user: `/actors/{uuid}/liked`

### Loading an item's comments

Loads the item's Context property (which can be it's top level item, or itself) and loads the `/objects/{uuid}/replies`

### Loading an item's votes

Loads the item's like collection `/objects/{uuid}/likes`
