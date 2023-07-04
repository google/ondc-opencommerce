// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package model contains schemas for request/response of ONDC Core API
//
// Most schema definitions are generated with [OpenAPI Generator]. The OpenAPI specification used
// to generate schema is [Retail API v1.0.30]. The generator option is [go-server].
//
// [OpenAPI Generator]: https://openapi-generator.tech
// [Retail API v1.0.30]: https://app.swaggerhub.com/apis/ONDC/ONDC-Protocol-Retail/1.0.30
// [go-server]: https://openapi-generator.tech/docs/generators/go-server
package model

import (
	"encoding/json"
	"time"
)

// AckResponse contains acknowledgement of the request.
type AckResponse struct {
	Message *MessageAck `json:"message,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// MessageAck is an inner message of AckResponse.
type MessageAck struct {
	Ack *Ack `json:"ack" validate:"required"`
}

// Ack - Describes the ACK response
type Ack struct {
	// Describe the status of the ACK response. If schema validation passes, status is ACK else it is NACK
	Status string `json:"status" validate:"oneof=ACK NACK"`

	// A list of tags containing any additional information sent along with the Acknowledgement.
	Tags []TagGroup `json:"tags,omitempty"`
}

// AddOn - Describes an add-on
type AddOn struct {
	// ID of the add-on. This follows the syntax {item.id}/add-on/{add-on unique id} for item specific add-on OR
	ID string `json:"id,omitempty"`

	Descriptor *Descriptor `json:"descriptor,omitempty"`

	Price *Price `json:"price,omitempty"`
}

// Address - Describes an address
type Address struct {
	// Door / Shop number of the address
	Door string `json:"door,omitempty"`

	// Name of address if applicable. Example, shop name
	Name string `json:"name,omitempty"`

	// Name of the building or block
	Building string `json:"building,omitempty"`

	// Street name or number
	Street string `json:"street,omitempty"`

	// Name of the locality, apartments
	Locality string `json:"locality,omitempty"`

	// Name or number of the ward if applicable
	Ward string `json:"ward,omitempty"`

	// City name
	City string `json:"city,omitempty"`

	// State name
	State string `json:"state,omitempty"`

	// Country name
	Country string `json:"country,omitempty"`

	// Area code. This can be Pincode, ZIP code or any equivalent
	AreaCode string `json:"area_code,omitempty"`
}

// Agent - Describes an order executor
type Agent struct {
	Person
	Contact

	// Since both `Person` and `contact` struct have `Tags` field,
	// we need to add this `Tags` field to avoid ambiguity.
	Tags *TagGroup `json:"tags,omitempty"`

	Rateable *Rateable `json:"rateable,omitempty"`
}

// Authorization - Describes an authorization mechanism
type Authorization struct {
	// Type of authorization mechanism used
	Type string `json:"type,omitempty"`

	// Token used for authorization
	Token string `json:"token,omitempty"`

	// Timestamp in RFC3339 format from which token is valid
	ValidFrom time.Time `json:"valid_from,omitempty"`

	// Timestamp in RFC3339 format until which token is valid
	ValidTo time.Time `json:"valid_to,omitempty"`

	// Status of the token
	Status string `json:"status,omitempty"`
}

// Billing - Describes a billing event
type Billing struct {
	// Personal details of the customer needed for billing.
	Name string `json:"name,omitempty"`

	Organization *Organization `json:"organization,omitempty"`

	Address *Address `json:"address,omitempty"`

	Email string `json:"email,omitempty"`

	Phone string `json:"phone,omitempty"`

	Time *Time `json:"time,omitempty"`

	// GST number
	TaxNumber string `json:"tax_number,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`

	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// Cancellation - Describes a cancellation event
type Cancellation struct {
	Type *string `json:"type,omitempty" validate:"omitempty,oneof=full partial"`

	RefID string `json:"ref_id,omitempty"`

	Policies []Policy `json:"policies,omitempty"`

	Time time.Time `json:"time,omitempty"`

	CancelledBy string `json:"cancelled_by,omitempty"`

	Reasons *Option `json:"reasons,omitempty"`

	SelectedReason *struct {
		ID string `json:"id,omitempty"`
	} `json:"selected_reason,omitempty"`

	AdditionalDescription *Descriptor `json:"additional_description,omitempty"`
}

// CancellationTerm - Describes the cancellation terms of an item or an order.
//
// This can be referenced at an item or order level. Item-level cancellation terms can override the terms at the order level.
type CancellationTerm struct {
	// Indicates whether a reason is required to cancel the order
	ReasonRequired bool `json:"reason_required,omitempty"`

	// Indicates if cancellation will result in a refund
	RefundEligible bool `json:"refund_eligible,omitempty"`

	// Indicates if cancellation will result in a return to origin
	ReturnEligible bool `json:"return_eligible,omitempty"`

	// The state of fulfillment during which these terms are applicable.
	FulfillmentState *struct {
		State
	} `json:"fulfillment_state,omitempty"`

	// Describes the return policy of an item or an order
	ReturnPolicy *struct {

		// Indicates if cancellation will result in a return to origin
		ReturnEligible bool `json:"return_eligible,omitempty"`

		// Applicable only for buyer managed returns where the buyer has to return the item to the origin before a certain date-time,
		// failing which they will not be eligible for refund.
		ReturnWithin *struct {
			Time
		} `json:"return_within,omitempty"`

		ReturnLocation       *Location `json:"return_location,omitempty"`
		FulfillmentManagedBy *string   `json:"fulfillment_managed_by,omitempty" validate:"omitempty,oneof=customer provider"`
	} `json:"return_policy,omitempty"`

	RefundPolicy *struct {
		// Indicates if cancellation will result in a refund
		RefundEligible bool `json:"refund_eligible,omitempty"`

		// Time within which refund will be processed after successful cancellation.
		RefundWithin *struct {
			Time
		} `json:"refund_within,omitempty"`

		RefundAmount *Price `json:"refund_amount,omitempty"`
	} `json:"refund_policy,omitempty"`

	// Information related to the time of cancellation.
	CancelBy *struct {
		Time
	} `json:"cancel_by,omitempty"`

	CancellationFee *Fee            `json:"cancellation_fee,omitempty"`
	XInputRequired  *XInput         `json:"xinput_required,omitempty"`
	XInputResponse  *XInputResponse `json:"xinput_response,omitempty"`
	ExternalRef     *MediaFile      `json:"external_ref,omitempty"`
}

// Catalog - Describes a Seller App catalog
type Catalog struct {
	BppDescriptor   *Descriptor   `json:"bpp/descriptor,omitempty"`
	BppCategories   []Category    `json:"bpp/categories,omitempty"`
	BppFulfillments []Fulfillment `json:"bpp/fulfillments,omitempty"`
	BppPayments     []Payment     `json:"bpp/payments,omitempty"`
	BppOffers       []Offer       `json:"bpp/offers,omitempty"`
	BppProviders    []Provider    `json:"bpp/providers,omitempty"`

	// Time after which catalog has to be refreshed
	Exp time.Time `json:"exp,omitempty"`
}

// Category - Describes a category
type Category struct {
	// Unique id of the category
	ID string `json:"id,omitempty"`

	// Unique id of the category
	ParentCategoryID string `json:"parent_category_id,omitempty"`

	Descriptor *Descriptor `json:"descriptor,omitempty"`

	Time *Time `json:"time,omitempty"`

	Tags *TagGroup `json:"tags,omitempty"`
}

// Circle - Describes a circular area on the map
type Circle struct {
	Gps *GPS `json:"gps" validate:"required"`

	Radius *Scalar `json:"radius" validate:"required"`
}

// City - Describes a city
type City struct {
	// Name of the city
	Name string `json:"name,omitempty"`

	// Codification of city code will be using the std code of the city e.g. for Bengaluru, city code is 'std:080'
	Code string `json:"code,omitempty"`
}

type Contact struct {
	Phone string `json:"phone,omitempty"`

	Email string `json:"email,omitempty"`

	Tags *TagGroup `json:"tags,omitempty"`
}

// Context - Describes a ONDC message context
type Context struct {
	Domain *Domain `json:"domain" validate:"required"`

	// Country code as per ISO 3166 Alpha-3 code format
	Country *string `json:"country" validate:"required"`

	// Codification of city code will be using the std code of the city e.g. for Bengaluru, city code is 'std:080'
	City *string `json:"city" validate:"required"`

	// Defines the ONDC API call. Any actions other than the enumerated actions are not supported by ONDC Protocol
	Action string `json:"action" validate:"oneof=search select init confirm update status track cancel rating support on_search on_select on_init on_confirm on_update on_status on_track on_cancel on_rating on_support"`

	// Version of ONDC core API specification being used
	CoreVersion *string `json:"core_version" validate:"required"`

	// Unique id of the Buyer App. By default it is the fully qualified domain name of the Buyer App
	BapID *string `json:"bap_id" validate:"required"`

	// URI of the Buyer App for accepting callbacks. Must have the same domain name as the bap_id
	BapURI *string `json:"bap_uri" validate:"required"`

	// Unique id of the Seller App. By default it is the fully qualified domain name of the Seller App,
	// mandatory for all peer-to-peer API requests, i.e. except search and on_search
	BppID string `json:"bpp_id,omitempty"`

	// URI of the Seller App. Must have the same domain name as the bap_id, mandatory for all
	// peer-to-peer API requests, i.e. except search and on_search
	BppURI string `json:"bpp_uri,omitempty"`

	// This is a unique value which persists across all API calls from search through confirm
	TransactionID *string `json:"transaction_id" validate:"required"`

	// This is a unique value which persists during a request / callback cycle
	MessageID *string `json:"message_id" validate:"required"`

	// Time of request generation in RFC3339 format
	Timestamp *time.Time `json:"timestamp" validate:"required"`

	// The encryption public key of the sender
	Key string `json:"key,omitempty"`

	// Timestamp for which this message holds valid in ISO8601 durations format -
	// Outer limit for TTL for search, select, init, confirm, status, track, cancel, update, rating, support is 'PT30S' which is 30 seconds,
	// different buyer apps can change this to meet their UX requirements, but it shouldn't exceed this outer limit
	TTL string `json:"ttl,omitempty"`
}

// Country - Describes a country.
type Country struct {
	// Name of the country
	Name string `json:"name,omitempty"`

	// Country code as per ISO 3166 Alpha-3 code format
	Code string `json:"code,omitempty"`
}

// Credential - Describes a credential of an entity - Person or Organization
type Credential struct {
	ID         string      `json:"id,omitempty"`
	Type       string      `json:"type,omitempty"` // TODO: handle default value: VerifiableCredential
	Descriptor *Descriptor `json:"descriptor,omitempty"`

	// URL of the credential
	URL string `json:"url,omitempty"`

	Tags *TagGroup `json:"tags,omitempty"`
}

// Descriptor - Describes the description of a real-world object.
type Descriptor struct {
	Name string `json:"name,omitempty"`

	Code string `json:"code,omitempty"`

	Symbol string `json:"symbol,omitempty"`

	ShortDesc string `json:"short_desc,omitempty"`

	LongDesc string `json:"long_desc,omitempty"`

	Images []Image `json:"images,omitempty"`

	Audio string `json:"audio,omitempty"`

	Var3dRender string `json:"3d_render,omitempty"`
}

// Dimensions - Describes the dimensions of a real-world object
type Dimensions struct {
	Length *Scalar `json:"length,omitempty"`

	Breadth *Scalar `json:"breadth,omitempty"`

	Height *Scalar `json:"height,omitempty"`
}

// Document - Describes a document which can be sent as a URL
type Document struct {
	URL string `json:"url,omitempty"`

	Label string `json:"label,omitempty"`
}

// Error - Describes an error object
type Error struct {
	Type string `json:"type" validate:"oneof=CONTEXT-ERROR CORE-ERROR DOMAIN-ERROR POLICY-ERROR JSON-SCHEMA-ERROR"`

	// ONDC specific error code. For full list of error codes, refer to docs/drafts/Error Codes.md of this repo
	Code *string `json:"code" validate:"required"`

	// Path to json schema generating the error. Used only during json schema validation errors
	Path string `json:"path,omitempty"`

	// Human readable message describing the error
	Message string `json:"message,omitempty"`
}

// Fee - A fee applied on a particular entity
type Fee struct {
	// Percentage of a value
	Percentage *struct {
		DecimalValue
	} `json:"percentage,omitempty"`

	// A fixed value
	Amount *struct {
		Price
	} `json:"amount,omitempty"`
}

// FeedbackFormElement - An element in the feedback form. It can be a question or an answer to the question.
type FeedbackFormElement struct {
	ID string `json:"id,omitempty"`

	ParentID string `json:"parent_id,omitempty"`

	// Specifies the question to which the answer options will be contained in the child FeedbackFormElements
	Question string `json:"question,omitempty"`

	// Specifies an answer option to which the question will be in the FeedbackFormElement specified in parent_id
	Answer string `json:"answer,omitempty"`

	// Specifies how the answer option should be rendered.
	AnswerType *string `json:"answer_type,omitempty" validate:"omitempty,oneof=radio checkbox text"`
}

// Feedback - Feedback for a service
type Feedback struct {
	FeedbackForm FeedbackForm `json:"feedback_form,omitempty"`

	FeedbackURL *FeedbackURL `json:"feedback_url,omitempty"`
}

// FeedbackURL - Describes how a feedback URL will be sent by the Seller App
type FeedbackURL struct {
	// feedback URL sent by the Seller App
	URL string `json:"url,omitempty"`

	TlMethod *string `json:"tl_method,omitempty" validate:"omitempty,oneof=http/get http/post"`

	Params *feedbackURLParams `json:"params,omitempty"`
}

type feedbackURLParams struct {
	// This value will be placed in the the $feedback_id url param in case of http/get and in the requestBody http/post requests
	FeedbackID *string `json:"feedback_id" validate:"required"`
}

// Form - Describes a form
type Form struct {
	// The URL from where the form can be fetched.
	//
	// The content fetched from the url must be processed as per the mime_type specified in this object.
	// Once fetched, the rendering platform can choosed to render the form as-is as an embeddable element; or process it further to blend with the theme of the application.
	// In case the interface is non-visual, the the render can process the form data and reproduce it as per the standard specified in the form.
	URL string `json:"url,omitempty"`

	// The form content string.
	//
	// This content will again follow the mime_type field for processing. Typically forms should be sent as an html string starting with <form></form> tags.
	// The application must render this form after removing any css or javascript code if necessary.
	// The `action` attribute in the form should have a url where the form needs to be submitted.
	Data string `json:"data,omitempty"`

	// This field indicates the nature and format of the form received by querying the url.
	//
	// MIME types are defined and standardized in IETF's RFC 6838.
	MimeType string `json:"mime_type,omitempty"`
}

type fulfillmentCustomer struct {
	Person *Person `json:"person,omitempty"`

	Contact *Contact `json:"contact,omitempty"`
}

// FulfillmentEnd - Details on the end of fulfillment
type FulfillmentEnd struct {
	Location *Location `json:"location,omitempty"`

	Time *Time `json:"time,omitempty"`

	Instructions *Descriptor `json:"instructions,omitempty"`

	Contact *Contact `json:"contact,omitempty"`

	Person *Person `json:"person,omitempty"`

	Authorization *Authorization `json:"authorization,omitempty"`
}

// Fulfillment - Describes how a single product/service will be rendered/fulfilled to the end customer
type Fulfillment struct {
	// Unique reference ID to the fulfillment of an order
	ID string `json:"id,omitempty"`

	// This describes the type of fulfillment
	//
	// "Pickup" - Buyer picks up from store by themselves or through their logistics provider
	// "Delivery" - seller delivers to buyer
	Type string `json:"type" validate:"oneof=Delivery Pickup 'Delivery and Pickup' 'Reverse QC'"`

	// Fulfillment Category
	ONDCOrgCategory string `json:"@ondc/org/category,omitempty"`

	// Fulfillment turnaround time in ISO8601 durations format e.g. 'PT24H' indicates 24 hour TAT
	ONDCOrgTAT string `json:"@ondc/org/TAT,omitempty"`

	// ID of the provider
	ProviderID string `json:"provider_id,omitempty"`

	ONDCOrgProviderName string `json:"@ondc/org/provider_name,omitempty"`

	// Rating value given to the object
	Rating float32 `json:"rating,omitempty"`

	State *State `json:"state,omitempty"`

	// Indicates whether the fulfillment allows tracking
	Tracking bool `json:"tracking,omitempty"`

	Customer *fulfillmentCustomer `json:"customer,omitempty"`

	Agent *Agent `json:"agent,omitempty"`

	Person *Person `json:"person,omitempty"`

	Contact *Contact `json:"contact,omitempty"`

	Vehicle *Vehicle `json:"vehicle,omitempty"`

	Start *FulfillmentStart `json:"start,omitempty"`

	End *FulfillmentEnd `json:"end,omitempty"`

	Rateable *Rateable `json:"rateable,omitempty"`

	Tags *TagGroup `json:"tags,omitempty"`
}

// FulfillmentStart - Details on the start of fulfillment
type FulfillmentStart struct {
	Location *Location `json:"location,omitempty"`

	Time *Time `json:"time,omitempty"`

	Instructions *Descriptor `json:"instructions,omitempty"`

	Contact *Contact `json:"contact,omitempty"`

	Person *Person `json:"person,omitempty"`

	Authorization *Authorization `json:"authorization,omitempty"`
}

// Intent - Intent of a user. Used for searching for services.
//
// Buyer App can set finder fee type in payment."@ondc/org/buyer_app_finder_fee_type"
// and amount in "@ondc/org/buyer_app_finder_fee_amount"
type Intent struct {
	Descriptor *Descriptor `json:"descriptor,omitempty"`

	Provider *Provider `json:"provider,omitempty"`

	Fulfillment *Fulfillment `json:"fulfillment,omitempty"`

	Payment *Payment `json:"payment,omitempty"`

	Category *Category `json:"category,omitempty"`

	Offer *Offer `json:"offer,omitempty"`

	Item *Item `json:"item,omitempty"`

	Tags *TagGroup `json:"tags,omitempty"`
}

// Item - Describes a product or a service offered to the end consumer by the provider
type Item struct {
	// This is the most unique identifier of a service item. An example of an Item ID could be the SKU of a product.
	ID string `json:"id,omitempty"`

	// This is the most unique identifier of a service item. An example of an Item ID could be the SKU of a product.
	ParentItemID string `json:"parent_item_id,omitempty"`

	Descriptor *Descriptor `json:"descriptor,omitempty"`

	Price *Price `json:"price,omitempty"`

	// Unique id of the category
	CategoryID string `json:"category_id,omitempty"`

	// Categories this item can be listed under
	CategoryIDs []string `json:"category_ids,omitempty"`

	// Unique reference ID to the fulfillment of an order
	FulfillmentID *string `json:"fulfillment_id,omitempty"`

	// Rating value given to the object
	Rating float32 `json:"rating,omitempty"`

	LocationID string `json:"location_id,omitempty"`

	Time *Time `json:"time,omitempty"`

	Rateable *Rateable `json:"rateable,omitempty"`

	Matched bool `json:"matched,omitempty"`

	Related bool `json:"related,omitempty"`

	Recommended bool `json:"recommended,omitempty"`

	// whether the item is returnable
	ONDCOrgReturnable bool `json:"@ondc/org/returnable,omitempty"`

	// in case of return, whether the item should be picked up by seller
	ONDCOrgSellerPickupReturn bool `json:"@ondc/org/seller_pickup_return,omitempty"`

	// return window for the item in ISO8601 durations format e.g. 'PT24H' indicates 24 hour return window. Mandatory if \"@ondc/org/returnable\" is \"true\"
	ONDCOrgReturnWindow string `json:"@ondc/org/return_window,omitempty"`

	// whether the item is cancellable
	ONDCOrgCancellable bool `json:"@ondc/org/cancellable,omitempty"`

	// time from order confirmation by which item ready to ship in ISO8601 durations format (e.g. 'PT30M' indicates item ready to ship in 30 mins). Mandatory for category_id \"F&B\"
	ONDCOrgTimeToShip string `json:"@ondc/org/time_to_ship,omitempty"`

	// whether the catalog item is available on COD
	ONDCOrgAvailableOnCOD bool `json:"@ondc/org/available_on_cod,omitempty"`

	// contact details for consumer care
	ONDCOrgContactDetailsConsumerCare string `json:"@ondc/org/contact_details_consumer_care,omitempty"`

	// mandatory for category_id "Packaged Commodities"
	ONDCOrgStatutoryReqsPackagedCommodities *struct {
		// name of manufacturer or packer (in case manufacturer is not the packer) or name of importer for imported goods
		ManufacturerOrPackerName string `json:"manufacturer_or_packer_name,omitempty"`

		// address of manufacturer or packer (in case manufacturer is not the packer) or address of importer for imported goods
		ManufacturerOrPackerAddress string `json:"manufacturer_or_packer_address,omitempty"`

		// manufacturing license no
		MfgLicenseNo string `json:"mfg_license_no,omitempty"`

		// common or generic name of commodity
		CommonOrGenericNameOfCommodity string `json:"common_or_generic_name_of_commodity,omitempty"`

		// for packages with multiple products, the name and number of quantity of each (can be shown as \"name1-number_or_quantity; name2-number_or_quantity..\")
		MultipleProductsNameNumberOrQty string `json:"multiple_products_name_number_or_qty,omitempty"`

		// net quantity of commodity in terms of standard unit of weight or measure of commodity contained in package
		NetQuantityOrMeasureOfCommodityInPkg string `json:"net_quantity_or_measure_of_commodity_in_pkg,omitempty"`

		// month and year of manufacture or packing or import
		MonthYearOfManufacturePackingImport string `json:"month_year_of_manufacture_packing_import,omitempty"`

		// month and year of expiry
		ExpiryDate string `json:"expiry_date,omitempty"`
	} `json:"@ondc/org/statutory_reqs_packaged_commodities,omitempty"`

	// mandatory for category_id "Packaged food"
	ONDCOrgStatutoryReqsPrepackagedFood *struct {
		// list of ingredients (except single ingredient foods), can be shown as ingredient (with percentage); ingredient (with percentage);..) e.g. \"Puffed Rice (40%); Split Green Gram (20%); Ground Nuts (20%);..\"
		IngredientsInfo string `json:"ingredients_info,omitempty"`

		// nutritional info (can be shown as nutritional info (with unit, per standard unit, per serving);..) e.g. \"Energy(KCal) - (per 100kg) 420, (per serving 50g) 250; Protein(g) - (per 100kg) 12, (per serving 50g)6;..\"
		NutritionalInfo string `json:"nutritional_info,omitempty"`

		// food additives together with specific name or recognized International Numbering System (can be shown as additive1-name or number;additive2-name or number;..)
		AdditivesInfo string `json:"additives_info,omitempty"`

		// name of manufacturer or packer (for non-retail containers)
		ManufacturerOrPackerName string `json:"manufacturer_or_packer_name,omitempty"`

		// address of manufacturer or packer (for non-retail containers)
		ManufacturerOrPackerAddress string `json:"manufacturer_or_packer_address,omitempty"`

		// name of brand owner
		BrandOwnerName string `json:"brand_owner_name,omitempty"`

		// address of brand owner
		BrandOwnerAddress string `json:"brand_owner_address,omitempty"`

		// FSSAI logo of brand owner (url based image e.g. uri:http://path/to/image)
		BrandOwnerFSSAILogo string `json:"brand_owner_FSSAI_logo,omitempty"`

		// FSSAI license no of brand owner
		BrandOwnerFSSAILicenseNo string `json:"brand_owner_FSSAI_license_no,omitempty"`

		// FSSAI license no of manufacturer or marketer or packer or bottler if different from brand owner
		OtherFSSAILicenseNo string `json:"other_FSSAI_license_no,omitempty"`

		// net quantity
		NetQuantity string `json:"net_quantity,omitempty"`

		// name of importer
		ImporterName string `json:"importer_name,omitempty"`

		// address of importer
		ImporterAddress string `json:"importer_address,omitempty"`

		// FSSAI logo of importer (url based image e.g. uri:http://path/to/image)
		ImporterFSSAILogo string `json:"importer_FSSAI_logo,omitempty"`

		// FSSAI license no of importer
		ImporterFSSAILicenseNo string `json:"importer_FSSAI_license_no,omitempty"`

		// country of origin for imported products (ISO 3166 Alpha-3 code format)
		ImportedProductCountryOfOrigin string `json:"imported_product_country_of_origin,omitempty"`

		// name of importer for product manufactured outside but packaged or bottled in India
		OtherImporterName string `json:"other_importer_name,omitempty"`

		// address of importer for product manufactured outside but packaged or bottled in India
		OtherImporterAddress string `json:"other_importer_address,omitempty"`

		// premises where product manufactured outside are packaged or bottled in India
		OtherPremises string `json:"other_premises,omitempty"`
	} `json:"@ondc/org/statutory_reqs_prepackaged_food,omitempty"`

	Tags *TagGroup `json:"tags,omitempty"`
}

// ItemQuantity - Describes count or amount of an item
type ItemQuantity struct {
	Allocated *itemQuantityInner `json:"allocated,omitempty"`

	Available *itemQuantityInner `json:"available,omitempty"`

	Maximum *itemQuantityMaximum `json:"maximum,omitempty"`

	Minimum *itemQuantityInner `json:"minimum,omitempty"`

	Selected *itemQuantityInner `json:"selected,omitempty"`

	Unitized *itemQuantityInner `json:"unitized,omitempty"`
}

type itemQuantityInner struct {
	Count int32 `json:"count,omitempty"`

	Measure *Scalar `json:"measure,omitempty"`
}

type itemQuantityMaximum struct {
	Count int32 `json:"count,omitempty"`

	Measure *Scalar `json:"measure,omitempty"`
}

// Language - indicates language code. ONDC supports language codes as per ISO 639.2 standard
type Language struct {
	Code string `json:"code,omitempty"`
}

// Location - Describes the location of a runtime object.
type Location struct {
	ID string `json:"id,omitempty"`

	Descriptor *Descriptor `json:"descriptor,omitempty"`

	Gps *GPS `json:"gps,omitempty"`

	Address *Address `json:"address,omitempty"`

	StationCode string `json:"station_code,omitempty"`

	City *City `json:"city,omitempty"`

	Country *Country `json:"country,omitempty"`

	Circle *Circle `json:"circle,omitempty"`

	Polygon string `json:"polygon,omitempty"`

	Var3dspace string `json:"3dspace,omitempty"`

	Time *Time `json:"time,omitempty"`
}

// MediaFile - This object contains a url to a media file.
type MediaFile struct {
	// indicates the nature and format of the document, file, or assortment of bytes. MIME types are defined and standardized in IETF's RFC 6838
	MimeType string `json:"mimetype,omitempty"`

	// The URL of the file
	URL string `json:"url,omitempty"`

	// The digital signature of the file signed by the sender
	Signature string `json:"signature,omitempty"`

	// The signing algorithm used by the sender
	DSA string `json:"dsa,omitempty"`
}

// Offer - Describes an offer
type Offer struct {
	ID string `json:"id,omitempty"`

	Descriptor *Descriptor `json:"descriptor,omitempty"`

	LocationIds []string `json:"location_ids,omitempty"`

	CategoryIds []string `json:"category_ids,omitempty"`

	ItemIds []string `json:"item_ids,omitempty"`

	Time *Time `json:"time,omitempty"`

	Tags *TagGroup `json:"tags,omitempty"`
}

type operatorAllOfExperience struct {
	Label string `json:"label,omitempty"`

	Value string `json:"value,omitempty"`

	Unit string `json:"unit,omitempty"`
}

type operatorAllOf struct {
	Experience *operatorAllOfExperience `json:"experience,omitempty"`
}

// Operator - Describes the agent of a service
type Operator struct {
	// Describes the name of a person in format: ./{given_name}/{honorific_prefix}/{first_name}/{middle_name}/{last_name}/{honorific_suffix}
	Name string `json:"name,omitempty"`

	// Image of an object. <br/><br/> A url based image will look like <br/><br/>```uri:http://path/to/image``` <br/><br/> An image can also be sent as a data string. For example : <br/><br/> ```data:js87y34ilhriuho84r3i4```
	Image string `json:"image,omitempty"`

	Dob string `json:"dob,omitempty"`

	// Gender of something, typically a Person, but possibly also fictional characters, animals, etc. While Male and Female may be used, text strings are also acceptable for people who do not identify as a binary gender
	Gender string `json:"gender,omitempty"`

	Cred string `json:"cred,omitempty"`

	// Describes a tag. This is a simple key-value store which is used to contain extended metadata
	Tags map[string]string `json:"tags,omitempty"`

	Experience *operatorAllOfExperience `json:"experience,omitempty"`
}

// Option - Describes a selectable option
type Option struct {
	ID string `json:"id,omitempty"`

	Descriptor *Descriptor `json:"descriptor,omitempty"`
}

type orderAddOnsInner struct {
	// ID of the add-on. This follows the syntax {item.id}/add-on/{add-on unique id} for item specific add-on OR
	ID string `json:"id,omitempty"`
}

// Order - Describes the details of an order
type Order struct {
	// Unique identifier for Order across the network
	ID string `json:"id,omitempty"`

	State string `json:"state,omitempty"`

	Provider *orderProvider `json:"provider,omitempty"`

	Items []orderItemsInner `json:"items,omitempty"`

	AddOns []orderAddOnsInner `json:"add_ons,omitempty"`

	Offers []orderProviderLocationsInner `json:"offers,omitempty"`

	Documents []Document `json:"documents,omitempty"`

	Billing *Billing `json:"billing,omitempty"`

	Fulfillments []Fulfillment `json:"fulfillments,omitempty"`

	// The cancellation terms of this order. This can be overriden at the item level cancellation terms.
	CancellationTerms []CancellationTerm `json:"cancellation_terms,omitempty"`

	Quote *Quotation `json:"quote,omitempty"`

	Payment *Payment `json:"payment,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`

	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type orderItemsInner struct {
	// This is the most unique identifier of a service item. An example of an Item ID could be the SKU of a product.
	ID string `json:"id,omitempty"`

	Quantity *struct {
		Count   int32   `json:"count,omitempty"`
		Measure *Scalar `json:"measure,omitempty"`
	} `json:"quantity,omitempty"`
}

type orderProvider struct {
	// ID of the provider
	ID string `json:"id,omitempty"`

	Locations []orderProviderLocationsInner `json:"locations,omitempty"`
}

type orderProviderLocationsInner struct {
	ID string `json:"id,omitempty"`
}

// Organization - Describes an organization
type Organization struct {
	Name string `json:"name,omitempty"`

	Cred string `json:"cred,omitempty"`
}

// Page - Describes a page in a search result
type Page struct {
	ID string `json:"id,omitempty"`

	NextID string `json:"next_id,omitempty"`
}

// Payment - Describes a payment
type Payment struct {
	// A payment uri to be called by the Buyer App. If empty, then the payment is to be done offline. The details of payment should be present in the params object. If ```tl_method``` = http/get, then the payment details will be sent as url params. Two url param values, ```$transaction_id``` and ```$amount``` are mandatory. And example url would be : https://www.example.com/pay?txid=$transaction_id&amount=$amount&vpa=upiid&payee=shopez&billno=1234
	URI string `json:"uri,omitempty"`

	TlMethod *string `json:"tl_method,omitempty" validate:"omitempty,oneof=http/get http/post payto upi"`

	Params *paymentParams `json:"params,omitempty"`

	Type *string `json:"type,omitempty" validate:"omitempty,oneof=ON-ORDER PRE-FULFILLMENT ON-FULFILLMENT POST-FULFILLMENT"`

	Status *string `json:"status,omitempty" validate:"omitempty,oneof=PAID NOT-PAID"`

	Time *Time `json:"time,omitempty"`

	CollectedBy *string `json:"collected_by,omitempty" validate:"omitempty,oneof=BAP BPP"`

	ONDCOrgCollectedByStatus *string `json:"@ondc/org/collected_by_status,omitempty" validate:"omitempty,oneof=Assert Agree Disagree Terminate"`

	// NOTE: valid values from swagger are `Amount` and `Percent` but valid values from API contract are `amount` and `percent`.
	// So we support both form.
	ONDCOrgBuyerAppFinderFeeType *string `json:"@ondc/org/buyer_app_finder_fee_type,omitempty" validate:"omitempty,oneof=Amount Percent amount percent"`

	ONDCOrgBuyerAppFinderFeeAmount *DecimalValue `json:"@ondc/org/buyer_app_finder_fee_amount,omitempty"`

	ONDCOrgWithholdingAmount *DecimalValue `json:"@ondc/org/withholding_amount,omitempty"`

	ONDCOrgWithholdingAmountStatus *string `json:"@ondc/org/withholding_amount_status,omitempty" validate:"omitempty,oneof=Assert Agree Disagree Terminate"`

	// return window for withholding amount in ISO8601 durations format e.g. 'PT24H' indicates 24 hour return window
	ONDCOrgReturnWindow string `json:"@ondc/org/return_window,omitempty"`

	ONDCOrgReturnWindowStatus *string `json:"@ondc/org/return_window_status,omitempty" validate:"omitempty,oneof=Assert Agree Disagree Terminate"`

	// In case of prepaid payment, whether settlement between counterparties should be on the basis of collection, shipment or delivery
	ONDCOrgSettlementBasis *string `json:"@ondc/org/settlement_basis,omitempty" validate:"omitempty,oneof=shipment delivery return_window_expiry"`

	ONDCOrgSettlementBasisStatus *string `json:"@ondc/org/settlement_basis_status,omitempty" validate:"omitempty,oneof=Assert Agree Disagree Terminate"`

	// settlement window for the counterparty in ISO8601 durations format e.g. 'PT48H' indicates T+2 settlement
	ONDCOrgSettlementWindow string `json:"@ondc/org/settlement_window,omitempty"`

	ONDCOrgSettlementWindowStatus *string `json:"@ondc/org/settlement_window_status,omitempty" validate:"omitempty,oneof=Assert Agree Disagree Terminate"`

	ONDCOrgSettlementDetails []struct {
		SettlementCounterparty  *string   `json:"settlement_counterparty,omitempty" validate:"omitempty,oneof=buyer buyer-app seller-app logistics-provider"`
		SettlementPhase         *string   `json:"settlement_phase,omitempty" validate:"omitempty,oneof=sale-amount withholding-amount refund"`
		SettlementAmount        int       `json:"settlement_amount,omitempty"`
		SettlementType          *string   `json:"settlement_type,omitempty" validate:"omitempty,oneof=neft rtgs upi credit"`
		SettlementBankAccountNo string    `json:"settlement_bank_account_no,omitempty"`
		SettlementIFSCCode      string    `json:"settlement_ifsc_code,omitempty"`
		UPIAddress              string    `json:"upi_address,omitempty"`         // UPI payment address e.g. VPA
		BankName                string    `json:"bank_name,omitempty"`           // Bank name
		BranchName              string    `json:"branch_name,omitempty"`         // Branch name
		BeneficiaryName         string    `json:"beneficiary_name,omitempty"`    // Beneficiary Name
		BeneficiaryAddress      string    `json:"beneficiary_address,omitempty"` // Beneficiary Address
		SettlementStatus        *string   `json:"settlement_status,omitempty" validate:"omitempty,oneof=PAID NOT-PAID"`
		SettlementReference     string    `json:"settlement_reference,omitempty"` // Settlement transaction reference number
		SettlementTimestamp     time.Time `json:"settlement_timestamp,omitempty"` // Settlement transaction timestamp
	} `json:"@ondc/org/settlement_details,omitempty"`
}

type paymentParams struct {
	// This value will be placed in the the $transaction_id url param in case of http/get and in the requestBody http/post requests
	TransactionID string `json:"transaction_id,omitempty"`

	TransactionStatus string `json:"transaction_status,omitempty"`

	// Describes a decimal value
	Amount string `json:"amount,omitempty"`

	// ISO 4217 alphabetic currency code e.g. 'INR'
	Currency *string `json:"currency" validate:"required"`
}

// Person - Describes a person.
type Person struct {
	Name *Name `json:"name,omitempty"`

	Image *Image `json:"image,omitempty"`

	Dob string `json:"dob,omitempty"`

	// Gender of something, typically a Person, but possibly also fictional characters, animals, etc. While Male and Female may be used, text strings are also acceptable for people who do not identify as a binary gender
	Gender string `json:"gender,omitempty"`

	Tags *TagGroup `json:"tags,omitempty"`
}

// Policy - Describes a policy. Allows for domain extension.
type Policy struct {
	ID string `json:"id,omitempty"`

	Descriptor *Descriptor `json:"descriptor,omitempty"`

	ParentPolicyID string `json:"parent_policy_id,omitempty"`

	Time *Time `json:"time,omitempty"`
}

// Price - Describes the price of an item. Allows for domain extension.
type Price struct {
	// ISO 4217 alphabetic currency code e.g. 'INR'
	Currency string `json:"currency,omitempty"`

	Value          *DecimalValue `json:"value,omitempty"`
	EstimatedValue *DecimalValue `json:"estimated_value,omitempty"`
	ComputedValue  *DecimalValue `json:"computed_value,omitempty"`
	ListedValue    *DecimalValue `json:"listed_value,omitempty"`
	OfferedValue   *DecimalValue `json:"offered_value,omitempty"`
	MinimumValue   *DecimalValue `json:"minimum_value,omitempty"`
	MaximumValue   *DecimalValue `json:"maximum_value,omitempty"`
}

// Provider - Describes a service provider. This can be a restaurant, a hospital, a Store etc
type Provider struct {
	// ID of the provider
	ID string `json:"id,omitempty"`

	Descriptor *Descriptor `json:"descriptor,omitempty"`

	// Category ID of the provider
	CategoryID string `json:"category_id,omitempty"`

	// FSSAI license no. Mandatory for category_id \"F&B\"
	ONDCOrgFSSAILicenseNo string `json:"@ondc/org/fssai_license_no,omitempty"`

	// Rating value given to the object
	Rating float32 `json:"rating,omitempty"`

	Time *Time `json:"time,omitempty"`

	Categories []Category `json:"categories,omitempty"`

	Cred []Credential `json:"creds,omitempty"`

	Fulfillments []Fulfillment `json:"fulfillments,omitempty"`

	Payments []Payment `json:"payments,omitempty"`

	Locations []providerLocationsInner `json:"locations,omitempty"`

	Offers []Offer `json:"offers,omitempty"`

	Items []Item `json:"items,omitempty"`

	// Validity of catalog in ISO8601 durations format after which it has to be refreshed
	//
	// e.g. 'P7D' indicates validity of 7 days; value of 0 indicates catalog is not cacheable
	TTL string `json:"ttl,omitempty"`

	// Time after which catalog has to be refreshed
	Exp time.Time `json:"exp,omitempty"`

	Rateable *Rateable `json:"rateable,omitempty"`

	Tags *TagGroup `json:"tags,omitempty"`
}

type providerLocationsInner struct {
	ID string `json:"id,omitempty"`

	Descriptor *Descriptor `json:"descriptor,omitempty"`

	// Describes a gps coordinate
	Gps string `json:"gps,omitempty"`

	Address *Address `json:"address,omitempty"`

	StationCode string `json:"station_code,omitempty"`

	City *City `json:"city,omitempty"`

	Country *Country `json:"country,omitempty"`

	Circle *Circle `json:"circle,omitempty"`

	Polygon string `json:"polygon,omitempty"`

	Var3dspace string `json:"3dspace,omitempty"`

	Time *Time `json:"time,omitempty"`

	Rateable *Rateable `json:"rateable,omitempty"`
}

type quotationBreakupInner struct {
	// This is the most unique identifier of a service item. An example of an Item ID could be the SKU of a product.
	ONDCOrgItemID string `json:"@ondc/org/item_id,omitempty"`

	ONDCOrgItemQuantity *itemQuantityInner `json:"@ondc/org/item_quantity,omitempty"`

	ONDCOrgTitleType *string `json:"@ondc/org/title_type,omitempty" validate:"omitempty,oneof=item delivery packing tax misc discount"`

	Item *Item `json:"item,omitempty"`

	Title string `json:"title,omitempty"`

	Price *Price `json:"price,omitempty"`
}

// Quotation - Describes a quote
type Quotation struct {
	Price *Price `json:"price,omitempty"`

	Breakup []quotationBreakupInner `json:"breakup,omitempty"`

	TTL *Duration `json:"ttl,omitempty"`
}

type ratingAck struct {
	// If feedback has been recorded or not
	FeedbackAck bool `json:"feedback_ack,omitempty"`

	// If rating has been recorded or not
	RatingAck bool `json:"rating_ack,omitempty"`
}

// Rating - Describes the rating of a person or an object.
type Rating struct {
	// Category of the object being rated
	RatingCategory string `json:"rating_category,omitempty"`

	// ID of the object being rated
	ID string `json:"id,omitempty"`

	// Rating value given to the object (1 - Poor; 2 - Needs improvement; 3 - Satisfactory; 4 - Good; 5 - Excellent)
	Value float32 `json:"value,omitempty" validate:"min=1,max=5"`

	FeedbackForm FeedbackForm `json:"feedback_form,omitempty"`

	// This value will be placed in the the $feedback_id url param in case of http/get and in the requestBody http/post requests
	FeedbackID string `json:"feedback_id,omitempty"`
}

// Scalar - An object representing a scalar quantity.
type Scalar struct {
	Type *string `json:"type,omitempty" validate:"omitempty,oneof=CONSTANT VARIABLE"`

	Value *float32 `json:"value" validate:"required"`

	EstimatedValue float32 `json:"estimated_value,omitempty"`

	ComputedValue float32 `json:"computed_value,omitempty"`

	Range *scalarRange `json:"range,omitempty"`

	Unit *string `json:"unit" validate:"required"`
}

type scalarRange struct {
	Min float32 `json:"min,omitempty"`

	Max float32 `json:"max,omitempty"`
}

// Schedule - Describes a schedule
type Schedule struct {
	Frequency *Duration `json:"frequency,omitempty"`

	Holidays []string `json:"holidays,omitempty"`

	Times []string `json:"times,omitempty"`
}

// State - Describes a state
type State struct {
	Descriptor *Descriptor `json:"descriptor,omitempty"`

	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// ID of entity which changed the state
	UpdatedBy string `json:"updated_by,omitempty"`
}

// Subscriber - Any entity which wants to authenticate itself on a network. This can be a Buyer App, Seller App or Gateway.
type Subscriber struct {
	// Registered domain name of the subscriber. Must have a valid SSL certificate issued by a Certificate Authority of the operating region
	SubscriberID string `json:"subscriber_id,omitempty"`

	Type *string `json:"type,omitempty" validate:"omitempty,oneof=bap bpp bg"`

	// Callback URL of the subscriber. The Registry will call this URL's on_subscribe API to validate the subscriber\\'s credentials
	CbURL string `json:"cb_url,omitempty"`

	Domain *Domain `json:"domain,omitempty"`

	// Codification of city code will be using the std code of the city e.g. for Bengaluru, city code is 'std:080'
	City string `json:"city,omitempty"`

	// Country code as per ISO 3166 Alpha-3 code format
	Country string `json:"country,omitempty"`

	// Signing Public key of the subscriber. <br/><br/>Any subscriber platform (Buyer App, Seller App, Gateway) who wants to transact on the network must digitally sign the ```requestBody``` using the corresponding private key of this public key and send it in the transport layer header. In case of ```HTTP``` it is the ```Authorization``` header. <br><br/>The ```Authorization``` will be used to validate the signature of a Buyer App or Seller App.<br/><br/>Furthermore, if an API call is being proxied or multicast by a ONDC Gateway, the Gateway must use it\\'s signing key to digitally sign the ```requestBody``` using the corresponding private key of this public key and send it in the ```X-Gateway-Authorization``` header.
	SigningPublicKey string `json:"signing_public_key,omitempty"`

	// Encryption public key of the Buyer App. Any Seller App must encrypt the ```requestBody.message``` value of the ```on_search``` API using this public key.
	EncryptionPublicKey string `json:"encryption_public_key,omitempty"`

	Status *string `json:"status,omitempty" validate:"omitempty,oneof=INITIATED UNDER_SUBSCRIPTION SUBSCRIBED INVALID_SSL UNSUBSCRIBED"`

	// Timestamp when a subscriber was added to the registry with status = INITIATED
	Created time.Time `json:"created,omitempty"`

	Updated time.Time `json:"updated,omitempty"`

	// Expiry timestamp in UTC derived from the ```lease_time``` of the subscriber
	Expires time.Time `json:"expires,omitempty"`
}

// Support - Customer support
type Support struct {
	Type *string `json:"type,omitempty" validate:"omitempty,oneof=order billing fulfillment"`

	RefID string `json:"ref_id,omitempty"`

	Channels *TagGroup `json:"channels,omitempty"`
}

// Tag - Describes a tag.
//
// This is a simple key-value store which is used to contain extended metadata.
// This object can be added as a property to any schema to describe extended attributes.
// For BAPs, tags can be sent during search to optimize and filter search results.
// BPPs can use tags to index their catalog to allow better search functionality.
// Tags are sent by the BPP as part of the catalog response in the `on_search` callback.
// Tags are also meant for display purposes. Upon receiving a tag, BAPs are meant to render them as name-value pairs.
// This is particularly useful when rendering tabular information about a product or service.
type Tag struct {
	// The machine-readable name of the tag.
	//
	// The allowed values of this property can be published at three levels namely,
	// a) Core specification,
	// b) industry sector-specific adaptations, and
	// c) Network-specific adaptations.
	// Except core, each adaptation (sector or network) should prefix a unique namespace with the allowed value.
	Code string `json:"code,omitempty"`

	// The human-readable name of the tag. This set by the BPP and rendered as-is by the BAP.
	//
	// Sometimes, the network policy may reserve some names for this property. Values outside the reserved values can be set by the BPP.
	// However,the BAP may choose to rename or even ignore this value and render the output purely using the `code` property,
	// but it is recommended for BAPs to keep the name same to avoid confusion and provide consistency.
	Name string `json:"name,omitempty"`

	// The value of the tag. This set by the BPP and rendered as-is by the BAP.
	Value string `json:"value,omitempty"`

	// This value indicates if the tag is intended for display purposes.
	//
	// If set to `true`, then this tag must be displayed.
	// If it is set to `false`, it should not be displayed.
	// This value can override the group display value.
	Display bool `json:"display,omitempty"`
}

// TagGroup - A collection of tag objects with group level attributes.
//
// For detailed documentation on the Tags and Tag Groups schema go to https://github.com/beckn/protocol-specifications/discussions/316
type TagGroup struct {
	// Indicates the display properties of the tag group.
	//
	// If display is set to false, then the group will not be displayed.
	// If it is set to true, it should be displayed.
	// However, group-level display properties can be overriden by individual tag-level display property.
	// As this schema is purely for catalog display purposes, it is not recommended to send this value during search.
	Display bool `json:"display,omitempty"` // TODO: handle default value: true

	// The machine-readable name of the tag group.
	//
	// The allowed values of this property can be published at three levels namely,
	// a) Core specification,
	// b) industry sector-specific adaptations, and
	// c) Network-specific adaptations.
	// Except core, each adaptation (sector or network) should prefix a unique namespace with the allowed value.
	// Values outside the allowed values may or may not be ignored by the rendering platform.
	// As this schema is purely for catalog display purposes, it is not recommended to send this value during search.
	Code string `json:"code,omitempty"`

	// A human-readable string describing the heading under which the tags are to be displayed.
	//
	// Sometimes, the network policy may reserve some names for this property. Values outside the reserved values can be set by the BPP.
	// However,the BAP may choose to rename or even ignore this value and render the output purely using code property,
	// but it is recommended for BAPs to keep the name same to avoid confusion and provide consistency.
	// As this schema is purely for catalog display purposes, it is not recommended to send this value during `search`.
	Name string `json:"name,omitempty"`

	// An array of Tag objects listed under this group.
	//
	// This property can be set by BAPs during search to narrow the `search` and achieve more relevant results.
	// When received during `on_search`, BAPs must render this list under the heading described by the `name` property of this schema.
	List []Tag `json:"list,omitempty"`
}

// Time - Describes time in its various forms. It can be a single point in time; duration; or a structured timetable of operations
type Time struct {
	Label string `json:"label,omitempty"`

	Timestamp time.Time `json:"timestamp,omitempty"`

	Duration *Duration `json:"duration,omitempty"`

	Range *timeRange `json:"range,omitempty"`

	// comma separated values representing days of the week
	Days string `json:"days,omitempty"`

	Schedule *Schedule `json:"schedule,omitempty"`
}

type timeRange struct {
	Start string `json:"start,omitempty"`

	End string `json:"end,omitempty"`
}

// Tracking - Contains tracking information that can be used by the BAP to track the fulfillment of an order in real-time. which is useful for knowing the location of time sensitive deliveries.
type Tracking struct {
	// A unique tracking reference number
	ID string `json:"id,omitempty"`

	// A URL to the tracking endpoint.
	//
	// This can be a link to a tracking webpage, a webhook URL created by the BAP where BPP can push the tracking data, or a GET url creaed by the BPP which the BAP can poll to get the tracking data.
	// It can also be a websocket URL where the BPP can push real-time tracking data.
	URL string `json:"url,omitempty"`

	// In case there is no real-time tracking endpoint available, this field will contain the latest location of the entity being tracked. The BPP will update this value everytime the BAP calls the track API.
	Location *struct {
		Location
	} `json:"location,omitempty"`

	// This value indicates if the tracking is currently active or not.
	//
	// If this value is `active`, then the BAP can begin tracking the order.
	// If this value is `inactive`, the tracking URL is considered to be expired and the BAP should stop tracking the order.
	Status *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`

	Tags *TagGroup `json:"tags,omitempty"`
}

// Vehicle - Describes the properties of a vehicle used in a mobility service
type Vehicle struct {
	Category string `json:"category,omitempty"`

	Make string `json:"make,omitempty"`

	Model string `json:"model,omitempty"`

	Size string `json:"size,omitempty"`

	Variant string `json:"variant,omitempty"`

	Color string `json:"color,omitempty"`

	EnergyType string `json:"energy_type,omitempty"`

	Registration string `json:"registration,omitempty"`
}

// XInput - Contains any additional or extended inputs required to confirm an order.
//
// This is typically a Form Input. Sometimes, selection of catalog elements is not enough for the BPP to confirm an order.
// For example, to confirm a flight ticket, the airline requires details of the passengers along with information on baggage, identity, in addition to the class of ticket.
// Similarly, a logistics company may require details on the nature of shipment in order to confirm the shipping.
// A recruiting firm may require additional details on the applicant in order to confirm a job application.
// For all such purposes, the BPP can choose to send this object attached to any object in the catalog that is required to be sent while placing the order.
// This object can typically be sent at an item level or at the order level.
// The item level XInput will override the Order level XInput as it indicates a special requirement of information for that particular item.
// Hence the BAP must render a separate form for the Item and another form at the Order level before confirmation.
type XInput struct {
	Form
}

// XInputResponse - The response to the form fetched via the XInput URL
type XInputResponse []XInputResponseInner

// XInputResponse - The response to the form fetched via the XInput URL
type XInputResponseInner struct {
	// The _name_ attribute of the input tag in the XInput form
	Input string `json:"input,omitempty"`

	// The value of the input field. Files must be sent as data URLs.
	//
	// For more information on Data URLs visit https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/Data_URLs
	Value string `json:"value,omitempty"`
}

// Domain - Codification of domain for ONDC
type Domain struct {
	Value string `validate:"oneof=nic2004:52110 ONDC:RET10 ONDC:RET11 ONDC:RET12 ONDC:RET13 ONDC:RET14 ONDC:RET15 ONDC:RET16 ONDC:RET17 ONDC:RET18 ONDC:RET19"`
}

// UnmarshalJSON unmarshal underlying value
func (d *Domain) UnmarshalJSON(b []byte) error { return json.Unmarshal(b, &d.Value) }

// MarshalJSON marshal underlying value
func (d *Domain) MarshalJSON() ([]byte, error) { return json.Marshal(d.Value) }

// DecimalValue - Describes a decimal value
type DecimalValue struct {
	Value string `validate:"custom_decimal_value"`
}

// UnmarshalJSON unmarshal underlying value
func (d *DecimalValue) UnmarshalJSON(b []byte) error { return json.Unmarshal(b, &d.Value) }

// MarshalJSON marshal underlying value
func (d *DecimalValue) MarshalJSON() ([]byte, error) { return json.Marshal(d.Value) }

// Duration - Describes duration as per ISO8601 format
type Duration struct {
	Value string
}

// UnmarshalJSON unmarshal underlying value
func (d *Duration) UnmarshalJSON(b []byte) error { return json.Unmarshal(b, &d.Value) }

// MarshalJSON marshal underlying value
func (d *Duration) MarshalJSON() ([]byte, error) { return json.Marshal(d.Value) }

// FeedbackForm - Describes a feedback form that a Seller App can send to get feedback from the Buyer App
type FeedbackForm []FeedbackFormElement

// GPS - Describes a gps coordinate
type GPS struct {
	Value string `validate:"custom_gps"`
}

// UnmarshalJSON unmarshal underlying value
func (g *GPS) UnmarshalJSON(b []byte) error { return json.Unmarshal(b, &g.Value) }

// MarshalJSON marshal underlying value
func (g *GPS) MarshalJSON() ([]byte, error) { return json.Marshal(g.Value) }

// Image - Image of an object
//
// A url based image will look like
// `uri:http://path/to/image`
// image can also be sent as a data string. For example :
// `data:js87y34ilhriuho84r3i4`
type Image struct {
	Value string
}

// UnmarshalJSON unmarshal underlying value
func (i *Image) UnmarshalJSON(b []byte) error { return json.Unmarshal(b, &i.Value) }

// MarshalJSON marshal underlying value
func (i *Image) MarshalJSON() ([]byte, error) { return json.Marshal(i.Value) }

// Name - Describes the name of a person in format: ./{given_name}/{honorific_prefix}/{first_name}/{middle_name}/{last_name}/{honorific_suffix}
type Name struct {
	Value string `validate:"custom_name"`
}

// UnmarshalJSON unmarshal underlying value
func (n *Name) UnmarshalJSON(b []byte) error { return json.Unmarshal(b, &n.Value) }

// MarshalJSON marshal underlying value
func (n *Name) MarshalJSON() ([]byte, error) { return json.Marshal(n.Value) }

// Rateable - If the entity can be rated or not
type Rateable struct {
	Value bool
}

// UnmarshalJSON unmarshal underlying value
func (r *Rateable) UnmarshalJSON(b []byte) error { return json.Unmarshal(b, &r.Value) }

// MarshalJSON marshal underlying value
func (r *Rateable) MarshalJSON() ([]byte, error) { return json.Marshal(r.Value) }
