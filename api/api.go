package api

import (
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/dnaeon/go-ucs/version"
)

// The remote API endpoint to which we POST requests.
const apiEndpoint = "nuova"

// The UserAgent that we use for our requests.
const userAgent = "go-ucs/" + version.Version

// BaseResponse contains the base attributes as returned in a response from a
// Cisco UCS API endpoint.
type BaseResponse struct {
	Cookie           string `xml:"cookie,attr"`
	Response         string `xml:"response,attr"`
	ErrorCode        string `xml:"errorCode,attr,omitempty"`
	InvocationResult string `xml:"invocationResult,attr,omitempty"`
	ErrorDescription string `xml:"errorDescr,attr,omitempty"`
}

// IsError returns a boolean indicating whether the response contains errors.
func (b *BaseResponse) IsError() bool {
	return b.ErrorCode != ""
}

// Error implements the error interface.
func (b *BaseResponse) Error() string {
	return fmt.Sprintf("%s: %s (code %s)", b.ErrorDescription, b.InvocationResult, b.ErrorCode)
}

// ToError creates a new error.Error from the error response fields.
func (b *BaseResponse) ToError() error {
	return errors.New(b.Error())
}

// AaaLoginRequest is the type which is sent during initial login
// in order to obtain authentication cookie.
type AaaLoginRequest struct {
	XMLName    xml.Name `xml:"aaaLogin"`
	InName     string   `xml:"inName,attr"`
	InPassword string   `xml:"inPassword,attr"`
}

// AaaLoginResponse represents the response after a successful login to UCS manager.
type AaaLoginResponse struct {
	BaseResponse
	XMLName          xml.Name `xml:"aaaLogin"`
	OutCookie        string   `xml:"outCookie,attr,omitempty"`
	OutRefreshPeriod int      `xml:"outRefreshPeriod,attr,omitempty"`
	OutPriv          string   `xml:"outPriv,attr,omitempty"`
	OutDomains       string   `xml:"outDomains,attr,omitempty"`
	OutChannel       string   `xml:"outChannel,attr,omitempty"`
	OutEvtChannel    string   `xml:"outEvtChannel,attr,omitempty"`
	OutName          string   `xml:"outName,attr,omitempty"`
	OutVersion       string   `xml:"outVersion,attr,omitempty"`
	OutSessionId     string   `xml:"outSessionId,attr,omitempty"`
}

// AaaRefreshRequest type is used for sending a request to the remote API endpoint for
// refreshing a session using the 47-character cookie obtained from a previous refresh
// or during the initial authentication as returned in a AaaLoginResponse.
type AaaRefreshRequest struct {
	XMLName    xml.Name `xml:"aaaRefresh"`
	InName     string   `xml:"inName,attr"`
	InPassword string   `xml:"inPassword,attr"`
	InCookie   string   `xml:"inCookie,attr"`
}

// AaaRefreshResponse is the response associated to a AaaRefreshRequest.
type AaaRefreshResponse struct {
	XMLName xml.Name `xml:"aaaRefresh"`
	AaaLoginResponse
}

// AaaLogoutRequest type is used for sending a request to invalidate an existing
// authentication cookie.
type AaaLogoutRequest struct {
	XMLName  xml.Name `xml:"aaaLogout"`
	InCookie string   `xml:"inCookie,attr"`
}

// AaaLogoutResponse represents the type that is returned after a call to aaaLogout method.
type AaaLogoutResponse struct {
	BaseResponse
	XMLName   xml.Name `xml:"aaaLogout"`
	OutStatus string   `xml:"outStatus,attr,omitempty"`
}

// AaaKeepAliveRequest type is used for sending a request for keeping a session active
// until the default session time expires.
type AaaKeepAliveRequest struct {
	XMLName xml.Name `xml:"aaaKeepAlive"`
	Cookie  string   `xml:"cookie,attr"`
}

// AaaKeepAliveResponse is the response type associated with a AaaKeepAliveRequest.
type AaaKeepAliveResponse struct {
	BaseResponse
	XMLName xml.Name `xml:"aaaKeepAlive"`
	Cookie  string   `xml:"cookie,attr"`
}

// ConfigResolveDnRequest type is used for constructing requests that retrieve a
// single managed object with the given DN.
type ConfigResolveDnRequest struct {
	XMLName        xml.Name `xml:"configResolveDn"`
	Cookie         string   `xml:"cookie,attr"`
	Dn             string   `xml:"dn,attr"`
	InHierarchical string   `xml:"inHierarchical,attr,omitempty"`
}

// ConfigResolveDnResponse is the type associated with a ConfigResolveDnRequest type.
// Specific classes contained within OutConfig should be xml.Unmarshal'ed first.
type ConfigResolveDnResponse struct {
	BaseResponse
	XMLName   xml.Name `xml:"configResolveDn"`
	Dn        string   `xml:"dn,attr"`
	OutConfig InnerXml `xml:"outConfig"`
}

// Dn represents a single managed object DN.
type Dn struct {
	XMLName xml.Name `xml:"dn"`
	Value   string   `xml:"value,attr"`
}

// NewDn creates a new DN value.
func NewDn(value string) Dn {
	dn := Dn{
		Value: value,
	}

	return dn
}

// ConfigResolveDnsRequest type is used for constructing requests that retrieve
// managed objects for a list of given DNs.
type ConfigResolveDnsRequest struct {
	XMLName        xml.Name `xml:"configResolveDns"`
	Cookie         string   `xml:"cookie,attr"`
	InHierarchical string   `xml:"inHierarchical,attr,omitempty"`
	InDns          []Dn     `xml:"inDns>dn"`
}

// ConfigResolveDnsResponse is the response type associated with a ConfigResolveDnsRequest.
// The managed objects within OutConfigs field should be xml.Unmarshal'ed.
type ConfigResolveDnsResponse struct {
	BaseResponse
	XMLName       xml.Name `xml:"configResolveDns"`
	OutUnresolved []Dn     `xml:"outUnresolved>dn"`
	OutConfigs    InnerXml `xml:"outConfigs"`
}

// InnerXml represents a generic configuration retrieved by the various query methods.
// After a successful result from a query method a client should unmarshal the data
// contained within an InnerXml to the specific managed object.
type InnerXml struct {
	XMLName xml.Name
	Inner   []byte `xml:",innerxml"`
}

// ConfigResolveClassRequest type is used for constructing requests that retrieve
// managed objects of a given class.
type ConfigResolveClassRequest struct {
	XMLName        xml.Name  `xml:"configResolveClass"`
	Cookie         string    `xml:"cookie,attr"`
	ClassId        string    `xml:"classId,attr"`
	InHierarchical string    `xml:"inHierarchical,attr,omitempty"`
	InFilter       FilterAny `xml:"inFilter>any,omitempty"`
}

// ConfigResolveClassResponse is the type associated with a ConfigResolveClassRequest.
// Specific classes contained within OutConfigs should be xml.Unmarshal'ed first.
type ConfigResolveClassResponse struct {
	BaseResponse
	XMLName    xml.Name `xml:"configResolveClass"`
	OutConfigs InnerXml `xml:"outConfigs"`
}

// Id represents an ID of a class.
type Id struct {
	XMLName xml.Name `xml:"Id"`
	Value   string   `xml:"value,attr"`
}

// NewId creates a new class id.
func NewId(value string) Id {
	id := Id{
		Value: value,
	}

	return id
}

// ConfigResolveClassesRequest type is used for constructing requests that retrieve managed
// objects in several classes.
type ConfigResolveClassesRequest struct {
	XMLName        xml.Name `xml:"configResolveClasses"`
	Cookie         string   `xml:"cookie,attr"`
	InHierarchical string   `xml:"inHierarchical,attr,omitempty"`
	InIds          []Id     `xml:"inIds>Id"`
}

// ConfigResolveClassesResponse is the response type associated with a ConfigResolveClassesRequest.
type ConfigResolveClassesResponse struct {
	BaseResponse
	XMLName    xml.Name `xml:"configResolveClasses"`
	OutConfigs InnerXml `xml:"outConfigs"`
}

// ConfigResolveChildren type is used for constructing requests that retrieve
// children of managed objects under a specified DN. A filter can be used to
// reduce the number of children being returned.
type ConfigResolveChildrenRequest struct {
	XMLName        xml.Name  `xml:"configResolveChildren"`
	Cookie         string    `xml:"cookie,attr"`
	ClassId        string    `xml:"classId,attr"`
	InDn           string    `xml:"inDn,attr"`
	InHierarchical string    `xml:"inHierarchical,attr"`
	InFilter       FilterAny `xml:"inFilter>any,omitempty"`
}

// ConfigResolveChildrenResponse is the response type associated with a ConfigResolveChildrenRequest.
type ConfigResolveChildrenResponse struct {
	BaseResponse
	XMLName    xml.Name `xml:"configResolveChildren"`
	OutConfigs InnerXml `xml:"outConfigs"`
}

// FilterAny represents any valid filter.
type FilterAny interface{}

// FilterProperty represents a Property Filter.
type FilterProperty struct {
	Class    string `xml:"class,attr"`
	Property string `xml:"property,attr"`
	Value    string `xml:"value,attr"`
}

// FilterEq represents an Equality Filter.
type FilterEq struct {
	XMLName xml.Name `xml:"eq"`
	FilterProperty
}

// FilterNe represents a Not Equal Filter.
type FilterNe struct {
	XMLName xml.Name `xml:"ne"`
	FilterProperty
}

// FilterGt represents a Greater Than Filter.
type FilterGt struct {
	XMLName xml.Name `xml:"gt"`
	FilterProperty
}

// FilterGe represents a Greater Than Or Equal To Filter.
type FilterGe struct {
	XMLName xml.Name `xml:"ge"`
	FilterProperty
}

// FilterLt represents a Less Than Filter.
type FilterLt struct {
	XMLName xml.Name `xml:"lt"`
	FilterProperty
}

// FilterLe represents a Less Than Or Equal To Filter.
type FilterLe struct {
	XMLName xml.Name `xml:"le"`
	FilterProperty
}

// FilterWildcard represents a Wildcard Filter.
// The wildcard filter uses standard regular expression syntax.
type FilterWildcard struct {
	XMLName xml.Name `xml:"wcard"`
	FilterProperty
}

// FilterAnyBits represents an Any Bits Filter.
type FilterAnyBits struct {
	XMLName xml.Name `xml:"anybit"`
	FilterProperty
}

// FilterAllBits represents an All Bits Filter.
type FilterAllBits struct {
	XMLName xml.Name `xml:"allbits"`
	FilterProperty
}

// FilterAnd represents a composite AND Filter.
type FilterAnd struct {
	XMLName xml.Name `xml:"and"`
	Filters []FilterAny
}

// FilterOr represents a composite OR Filter.
type FilterOr struct {
	XMLName xml.Name `xml:"or"`
	Filters []FilterAny
}

// FilterNot represents a NOT Modifier Filter.
type FilterNot struct {
	XMLName xml.Name `xml:"not"`
	Filters []FilterAny
}

// FilterBetween represents a Between Filter.
type FilterBetween struct {
	XMLName     xml.Name `xml:"bw"`
	Class       string   `xml:"class,attr"`
	Property    string   `xml:"property,attr"`
	FirstVault  string   `xml:"firstValue,attr"`
	SecondValue string   `xml:"secondValue,attr"`
}
