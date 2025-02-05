package types

import (
	"strconv"
	"strings"
)

const (
	// Values for FilterAttrs.Protocol
	FilterProtocolAll = "all"
	FilterProtocolIP  = "ip" // note ip == ipv4

	// FlowerFilter.Kind
	FilterKindFlower = "flower"

	// FlowerKeys
	FlowerKeyIPProto = "ip_proto"
	FlowerKeyDstIP   = "dst_ip"
	FlowerKeyDstPort = "dst_port"
)

// Filter represent a tc filter object
type Filter interface {
	CmdLineGenerator
	// Attrs returns FilterAttrs
	Attrs() *FilterAttrs
	// Equals compares this Filter with other, returns true if they are equal or false otherwise
	Equals(other Filter) bool
}

// FilterAttrs holds filter object attributes
type FilterAttrs struct {
	Kind     string
	Protocol string
	Chain    *uint16
	Handle   *uint32
	Priority *uint16
}

// NewFilterAttrs creates new FilterAttrs instance
func NewFilterAttrs(kind string, protocol string, chain *uint16, handle *uint32, priority *uint16) *FilterAttrs {
	return &FilterAttrs{
		Kind:     kind,
		Protocol: protocol,
		Chain:    chain,
		Handle:   handle,
		Priority: priority,
	}
}

// GenCmdLineArgs implements CmdLineGenerator interface, it generates the needed tc command line args for FilterAttrs
func (fa *FilterAttrs) GenCmdLineArgs() []string {
	args := []string{}

	if fa.Protocol != "" {
		args = append(args, "protocol", fa.Protocol)
	}

	if fa.Handle != nil {
		args = append(args, "handle", strconv.FormatUint(uint64(*fa.Handle), 10))
	}

	if fa.Chain != nil {
		args = append(args, "chain", strconv.FormatUint(uint64(*fa.Chain), 10))
	}

	if fa.Priority != nil {
		args = append(args, "pref", strconv.FormatUint(uint64(*fa.Priority), 10))
	}

	// must be last as next are filter type specific params
	args = append(args, fa.Kind)

	return args
}

// Equals compares this FilterAttrs with other, returns true if they are equal or false otherwise
func (fa *FilterAttrs) Equals(other *FilterAttrs) bool {
	if fa == other {
		return true
	}

	if (fa == nil && other != nil) || (fa != nil && other == nil) {
		return false
	}

	if fa.Kind != other.Kind {
		return false
	}
	if fa.Protocol != other.Protocol {
		return false
	}
	defChain := uint16(ChainDefaultChain)
	if !compare(fa.Chain, other.Chain, &defChain) {
		return false
	}
	if !compare(fa.Priority, other.Priority, nil) {
		return false
	}
	return true
}

// FlowerSpec holds flower filter specification (which consists of a list of Match)
type FlowerSpec struct {
	IpProto *string
	DstIP   *string
	DstPort *uint16
}

// GenCmdLineArgs implements CmdLineGenerator interface, it generates the needed tc command line args for FlowerSpec
func (ff *FlowerSpec) GenCmdLineArgs() []string {
	args := []string{}

	if ff == nil {
		return args
	}

	if ff.IpProto != nil {
		args = append(args, FlowerKeyIPProto, *ff.IpProto)
	}

	if ff.DstIP != nil {
		args = append(args, FlowerKeyDstIP, *ff.DstIP)
	}

	if ff.DstPort != nil {
		args = append(args, FlowerKeyDstPort, strconv.FormatUint(uint64(*ff.DstPort), 10))
	}

	return args
}

// Equals compares this FlowerSpec with other, returns true if they are equal or false otherwise
func (ff *FlowerSpec) Equals(other *FlowerSpec) bool {
	if ff == other {
		return true
	}

	if (ff == nil && other != nil) || (ff != nil && other == nil) {
		return false
	}

	// same Key/val
	if !compare(ff.IpProto, other.IpProto, nil) {
		return false
	}
	if !compare(ff.DstIP, other.DstIP, nil) {
		return false
	}
	if !compare(ff.DstPort, other.DstPort, nil) {
		return false
	}

	return true
}

// FlowerFilter is a concrete implementation of Filter of kind Flower
type FlowerFilter struct {
	FilterAttrs
	// Flower Match keys, only valid if Kind == FilterKindFlower
	Flower *FlowerSpec
	// Actions
	Actions []Action
}

// Attrs implements Filter interface, it returns FilterAttrs
func (f *FlowerFilter) Attrs() *FilterAttrs {
	return &f.FilterAttrs
}

// Equals implements Filter interface
func (f *FlowerFilter) Equals(other Filter) bool {
	// types equal
	otherFlower, ok := other.(*FlowerFilter)
	if !ok {
		return false
	}

	// FilterAttr equal
	if !f.Attrs().Equals(other.Attrs()) {
		return false
	}

	// FlowerSpec Equal
	if !f.Flower.Equals(otherFlower.Flower) {
		return false
	}

	// Actions Equal (order matters)
	if len(f.Actions) != len(otherFlower.Actions) {
		return false
	}
	for i := range f.Actions {
		if !f.Actions[i].Equals(otherFlower.Actions[i]) {
			return false
		}
	}

	return true
}

// GenCmdLineArgs implements CmdLineGenerator interface, it generates the needed tc command line args for FlowerFilter
func (f *FlowerFilter) GenCmdLineArgs() []string {
	args := []string{}

	args = append(args, f.FilterAttrs.GenCmdLineArgs()...)

	if f.Flower != nil {
		args = append(args, f.Flower.GenCmdLineArgs()...)
	}

	for _, action := range f.Actions {
		args = append(args, action.GenCmdLineArgs()...)
	}

	return args
}

// Builders

// NewFilterAttrsBuilder returns a new FilterAttrsBuilder
func NewFilterAttrsBuilder() *FilterAttrsBuilder {
	return &FilterAttrsBuilder{}
}

// FilterAttrsBuilder is a FilterAttr builder
type FilterAttrsBuilder struct {
	filterAttrs FilterAttrs
}

// WithKind adds Kind to FilterAttrsBuilder
func (fb *FilterAttrsBuilder) WithKind(k string) *FilterAttrsBuilder {
	fb.filterAttrs.Kind = k
	return fb
}

// WithProtocol adds Protocol to FilterAttrsBuilder
func (fb *FilterAttrsBuilder) WithProtocol(p string) *FilterAttrsBuilder {
	fb.filterAttrs.Protocol = p
	return fb
}

// WithChain adds Chain index to FilterAttrsBuilder
func (fb *FilterAttrsBuilder) WithChain(c uint16) *FilterAttrsBuilder {
	fb.filterAttrs.Chain = &c
	return fb
}

// WithHandle adds Handle to FilterAttrsBuilder
func (fb *FilterAttrsBuilder) WithHandle(h uint32) *FilterAttrsBuilder {
	fb.filterAttrs.Handle = &h
	return fb
}

// WithPriority adds Priority to FilterAttrsBuilder
func (fb *FilterAttrsBuilder) WithPriority(p uint16) *FilterAttrsBuilder {
	fb.filterAttrs.Priority = &p
	return fb
}

// Build builds and returns a new FilterAttrs instance
// Note: calling Build() multiple times will not return a completely
// new object on each call. that is, pointer/slice/map types will not be deep copied.
// to create several objects, different builders should be used.
func (fb *FilterAttrsBuilder) Build() *FilterAttrs {
	return NewFilterAttrs(fb.filterAttrs.Kind, fb.filterAttrs.Protocol, fb.filterAttrs.Chain, fb.filterAttrs.Handle, fb.filterAttrs.Priority)
}

// NewFlowerFilterBuilder returns a new instance of FlowerFilterBuilder
func NewFlowerFilterBuilder() *FlowerFilterBuilder {
	return &FlowerFilterBuilder{
		filterAttrsBuilder: NewFilterAttrsBuilder(),
		flowerFilter: FlowerFilter{
			Flower:  &FlowerSpec{},
			Actions: make([]Action, 0),
		},
	}
}

// FlowerFilterBuilder is a FlowerFilter builder
type FlowerFilterBuilder struct {
	filterAttrsBuilder *FilterAttrsBuilder
	flowerFilter       FlowerFilter
}

// WithKind adds Kind to FlowerFilterBuilder
func (fb *FlowerFilterBuilder) WithKind(k string) *FlowerFilterBuilder {
	fb.filterAttrsBuilder = fb.filterAttrsBuilder.WithKind(k)
	return fb
}

// WithProtocol adds Protocol to FlowerFilterBuilder
func (fb *FlowerFilterBuilder) WithProtocol(p string) *FlowerFilterBuilder {
	fb.filterAttrsBuilder = fb.filterAttrsBuilder.WithProtocol(p)
	return fb
}

// WithChain adds Chain number to FlowerFilterBuilder
func (fb *FlowerFilterBuilder) WithChain(c uint16) *FlowerFilterBuilder {
	fb.filterAttrsBuilder = fb.filterAttrsBuilder.WithChain(c)
	return fb
}

// WithHandle adds Handle to FlowerFilterBuilder
func (fb *FlowerFilterBuilder) WithHandle(h uint32) *FlowerFilterBuilder {
	fb.filterAttrsBuilder = fb.filterAttrsBuilder.WithHandle(h)
	return fb
}

// WithPriority adds Priority to FlowerFilterBuilder
func (fb *FlowerFilterBuilder) WithPriority(p uint16) *FlowerFilterBuilder {
	fb.filterAttrsBuilder = fb.filterAttrsBuilder.WithPriority(p)
	return fb
}

// WithMatchKeyIPProto adds Match with FlowerKeyIPProto key and specified value to FlowerFilterBuilder
func (fb *FlowerFilterBuilder) WithMatchKeyIPProto(val string) *FlowerFilterBuilder {
	lower := strings.ToLower(val)
	fb.flowerFilter.Flower.IpProto = &lower
	return fb
}

// WithMatchKeyDstIP adds Match with FlowerKeyDstIP key and specified value to FlowerFilterBuilder
func (fb *FlowerFilterBuilder) WithMatchKeyDstIP(val string) *FlowerFilterBuilder {
	// if its a CIDR with /32 take only the IP else take the full CIDR
	if strings.HasSuffix(val, "/32") {
		valNoMask := val[:len(val)-3]
		fb.flowerFilter.Flower.DstIP = &valNoMask
	} else {
		fb.flowerFilter.Flower.DstIP = &val
	}
	return fb
}

// WithMatchKeyDstPort adds Match with FlowerKeyDstPort key and specified value to FlowerFilterBuilder
func (fb *FlowerFilterBuilder) WithMatchKeyDstPort(val uint16) *FlowerFilterBuilder {
	fb.flowerFilter.Flower.DstPort = &val
	return fb
}

// WithAction adds specified Action to FlowerFilterBuilder
func (fb *FlowerFilterBuilder) WithAction(a Action) *FlowerFilterBuilder {
	fb.flowerFilter.Actions = append(fb.flowerFilter.Actions, a)
	return fb
}

// Build builds and creates a new FlowerFilter instance
// Note: calling Build() multiple times will not return a completely
// new object on each call. that is, pointer/slice/map types will not be deep copied.
// to create several objects, different builders should be used.
func (fb *FlowerFilterBuilder) Build() *FlowerFilter {
	fb.flowerFilter.FilterAttrs = *fb.filterAttrsBuilder.Build()
	fb.flowerFilter.Kind = FilterKindFlower

	return &FlowerFilter{
		FilterAttrs: *fb.flowerFilter.Attrs(),
		Flower:      fb.flowerFilter.Flower,
		Actions:     fb.flowerFilter.Actions,
	}
}
