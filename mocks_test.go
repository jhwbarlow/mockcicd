package main

import "time"

type mockObtainer struct {
	errorToReturn error

	obtainCalled bool
}

func newMockObtainer(errorToReturn error) *mockObtainer {
	return &mockObtainer{errorToReturn: errorToReturn}
}

func (mo *mockObtainer) Obtain(destPath string) error {
	mo.obtainCalled = true

	if mo.errorToReturn != nil {
		return mo.errorToReturn
	}

	return nil
}

type mockTagDeducer struct {
	tagToReturn string

	deduceCalled bool
	callCount    int
}

func newMockTagDeducer(tagToReturn string) *mockTagDeducer {
	return &mockTagDeducer{tagToReturn: tagToReturn}
}

func (md *mockTagDeducer) Deduce() (string, error) {
	md.deduceCalled = true
	md.callCount++

	return md.tagToReturn, nil
}

type mockCountingAsyncTagDeducer struct {
	closeAfterCallCount int
	tagToReturn         string
	errorToReturn       error
	deduceCountReached  chan<- struct{}
	deduceAcked         <-chan struct{}

	deduceCalled bool
	callCount    int
}

func newMockCountingAsyncTagDeducer(closeAfterCallCount int,
	tagToReturn string,
	errorToReturn error,
	deduceCountReached chan<- struct{},
	deduceAcked <-chan struct{}) *mockCountingAsyncTagDeducer {
	return &mockCountingAsyncTagDeducer{
		tagToReturn:         tagToReturn,
		errorToReturn:       errorToReturn,
		closeAfterCallCount: closeAfterCallCount,
		deduceCountReached:  deduceCountReached,
		deduceAcked:         deduceAcked,
	}
}

func (md *mockCountingAsyncTagDeducer) Deduce() (string, error) {
	md.deduceCalled = true
	md.callCount++

	if md.callCount == md.closeAfterCallCount {
		// As building will be the last step of the run() function if Deduce() returns an error,
		// we use Deduce() having been called as the signal to tell the test goroutine it is OK
		// to cancel the run() loop.
		close(md.deduceCountReached)

		// With fast poll periods, the run() goroutine will not yield the processor and hence the run()
		// loop will not have been cancelled by the test goroutine by the time Deduce() is called again.
		// This causes a panic as the md.deduceCountReached channel is already closed. So we handshake here to
		// yield to the test goroutine.
		<-md.deduceAcked
	}

	if md.errorToReturn != nil {
		return "", md.errorToReturn
	}

	return md.tagToReturn, nil
}

type mockBuilder struct {
	buildCalled bool
	callCount   int
}

func newMockBuilder() *mockBuilder {
	return new(mockBuilder)
}

func (mb *mockBuilder) Build(buildContextPath, name, tag string) error {
	mb.buildCalled = true
	mb.callCount++

	return nil
}

type mockCountingAsyncBuilder struct {
	closeAfterCallCount int
	errorToReturn       error
	buildCountReached   chan<- struct{}
	buildAcked          <-chan struct{}

	buildCalled bool
	callCount   int
}

func newMockCountingAsyncBuilder(closeAfterCallCount int,
	errorToReturn error,
	buildCountReached chan<- struct{},
	buildAcked <-chan struct{}) *mockCountingAsyncBuilder {
	return &mockCountingAsyncBuilder{
		errorToReturn:       errorToReturn,
		closeAfterCallCount: closeAfterCallCount,
		buildCountReached:   buildCountReached,
		buildAcked:          buildAcked,
	}
}

func (mb *mockCountingAsyncBuilder) Build(buildContextPath, name, tag string) error {
	mb.buildCalled = true
	mb.callCount++

	if mb.callCount == mb.closeAfterCallCount {
		// As building will be the last step of the run() function if Build() returns an error,
		// we use Build() having been called as the signal to tell the test goroutine it is OK
		// to cancel the run() loop.
		close(mb.buildCountReached)

		// With fast poll periods, the run() goroutine will not yield the processor and hence the run()
		// loop will not have been cancelled by the test goroutine by the time Build() is called again.
		// This causes a panic as the mb.buildCountReached channel is already closed. So we handshake here to
		// yield to the test goroutine.
		<-mb.buildAcked
	}

	if mb.errorToReturn != nil {
		return mb.errorToReturn
	}

	return nil
}

type mockPusher struct {
	pushCalled bool
	callCount  int
}

func newMockPusher() *mockPusher {
	return new(mockPusher)
}

func (mp *mockPusher) Push(name, tag string) error {
	mp.pushCalled = true
	mp.callCount++

	return nil
}

type mockCountingAsyncPusher struct {
	closeAfterCallCount int
	errorToReturn       error
	pushCountReached    chan<- struct{}
	pushAcked           <-chan struct{}

	pushCalled bool
	callCount  int
}

func newMockCountingAsyncPusher(closeAfterCallCount int,
	errorToReturn error,
	pushCountReached chan<- struct{},
	pushAcked <-chan struct{}) *mockCountingAsyncPusher {
	return &mockCountingAsyncPusher{
		errorToReturn:       errorToReturn,
		closeAfterCallCount: closeAfterCallCount,
		pushCountReached:    pushCountReached,
		pushAcked:           pushAcked,
	}
}

func (mp *mockCountingAsyncPusher) Push(name, tag string) error {
	mp.pushCalled = true
	mp.callCount++

	if mp.callCount == mp.closeAfterCallCount {
		// As pushing will be the last step of the run() function if Push() returns an error,
		// we use Push() having been called as the signal to tell the test goroutine it is OK
		// to cancel the run() loop.
		close(mp.pushCountReached)

		// With fast poll periods, the run() goroutine will not yield the processor and hence the run()
		// loop will not have been cancelled by the test goroutine by the time Push() is called again.
		// This causes a panic as the mp.pushCountReached channel is already closed. So we handshake here to
		// yield to the test goroutine.
		<-mp.pushAcked
	}

	if mp.errorToReturn != nil {
		return mp.errorToReturn
	}

	return nil
}

type mockInstaller struct {
	installCalled bool
	callCount     int
}

func newMockInstaller() *mockInstaller {
	return new(mockInstaller)
}

func (mi *mockInstaller) Install(imageName, imageTag string, timeout time.Duration) error {
	mi.installCalled = true
	mi.callCount++

	return nil
}

type mockAsyncInstaller struct {
	installed    chan<- struct{}
	installAcked <-chan struct{}

	installCalled bool
}

func newMockAsyncInstaller(installed chan<- struct{}, installAcked <-chan struct{}) *mockAsyncInstaller {
	return &mockAsyncInstaller{
		installed:    installed,
		installAcked: installAcked,
	}
}

func (mi *mockAsyncInstaller) Install(imageName, imageTag string, timeout time.Duration) error {
	mi.installCalled = true

	// As installing is the last step of the run() function, we use Install() having been called
	// as the signal to tell the test goroutine it is OK to cancel the run() loop.
	close(mi.installed)

	// With fast poll periods, the run() goroutine will not yield the processor and hence the run()
	// loop will not have been cancelled by the test goroutine by the time Install() is called again.
	// This causes a panic as the mi.installed channel is already closed. So we handshake here to
	// yield to the test goroutine.
	<-mi.installAcked

	return nil
}

type mockCountingAsyncInstaller struct {
	closeAfterCallCount int
	errorToReturn       error
	installCountReached chan<- struct{}
	installAcked        <-chan struct{}

	installCalled bool
	callCount     int
}

func newMockCountingAsyncInstaller(closeAfterCallCount int,
	errorToReturn error,
	installCountReached chan<- struct{},
	installAcked <-chan struct{}) *mockCountingAsyncInstaller {
	return &mockCountingAsyncInstaller{
		errorToReturn:       errorToReturn,
		closeAfterCallCount: closeAfterCallCount,
		installCountReached: installCountReached,
		installAcked:        installAcked,
	}
}

func (mi *mockCountingAsyncInstaller) Install(imageName, imageTag string, timeout time.Duration) error {
	mi.installCalled = true
	mi.callCount++

	if mi.callCount == mi.closeAfterCallCount {
		// As installing is the last step of the run() function, we use Install() having been called
		// as the signal to tell the test goroutine it is OK to cancel the run() loop.
		close(mi.installCountReached)

		// With fast poll periods, the run() goroutine will not yield the processor and hence the run()
		// loop will not have been cancelled by the test goroutine by the time Install() is called again.
		// This causes a panic as the mi.installCountReached channel is already closed. So we handshake here to
		// yield to the test goroutine.
		<-mi.installAcked
	}

	if mi.errorToReturn != nil {
		return mi.errorToReturn
	}

	return nil
}

type mockChecker struct {
	newVersionAvailable bool

	checkCalled bool
	callCount   int
}

func newMockChecker(newVersionAvailable bool) *mockChecker {
	return &mockChecker{newVersionAvailable: newVersionAvailable}
}

func (mc *mockChecker) Check() (bool, error) {
	mc.checkCalled = true
	mc.callCount++

	return mc.newVersionAvailable, nil
}

type mockCountingAsyncChecker struct {
	newVersionAvailable bool
	closeAfterCallCount int
	errorToReturn       error
	checkCountReached   chan<- struct{}
	checkAcked          <-chan struct{}

	checkCalled bool
	callCount   int
}

func newMockCountingAsyncChecker(closeAfterCallCount int,
	newVersionAvailable bool,
	errorToReturn error,
	checkCountReached chan<- struct{},
	checkAcked <-chan struct{}) *mockCountingAsyncChecker {
	return &mockCountingAsyncChecker{
		newVersionAvailable: newVersionAvailable,
		errorToReturn:       errorToReturn,
		closeAfterCallCount: closeAfterCallCount,
		checkCountReached:   checkCountReached,
		checkAcked:          checkAcked,
	}
}

func (mc *mockCountingAsyncChecker) Check() (bool, error) {
	mc.checkCalled = true
	mc.callCount++

	if mc.callCount == mc.closeAfterCallCount {
		// As checking will be the last step of the run() function if Check() returns an error,
		// we use Check() having been called as the signal to tell the test goroutine it is OK
		// to cancel the run() loop.
		close(mc.checkCountReached)

		// With fast poll periods, the run() goroutine will not yield the processor and hence the run()
		// loop will not have been cancelled by the test goroutine by the time Check() is called again.
		// This causes a panic as the mc.checkCountReached channel is already closed. So we handshake here to
		// yield to the test goroutine.
		<-mc.checkAcked
	}

	if mc.errorToReturn != nil {
		return false, mc.errorToReturn
	}

	return mc.newVersionAvailable, nil
}

type mockAsyncChecker struct {
	newVersionAvailable bool

	checked    chan<- struct{}
	checkAcked <-chan struct{}

	checkCalled bool
}

func newMockAsyncChecker(newVersionAvailable bool,
	checked chan<- struct{},
	checkAcked <-chan struct{}) *mockAsyncChecker {
	return &mockAsyncChecker{
		newVersionAvailable: newVersionAvailable,
		checked:             checked,
		checkAcked:          checkAcked,
	}
}

func (mc *mockAsyncChecker) Check() (bool, error) {
	mc.checkCalled = true

	// As checking is the last step of the run() function in the case where there is no change,
	// we use Check() having been called as the signal to tell the test goroutine it is OK to
	// cancel the run() loop.
	close(mc.checked)

	// With fast poll periods, the run() goroutine will not yield the processor and hence the run()
	// loop will not have been cancelled by the test goroutine by the time Check() is called again.
	// This causes a panic as the mc.checked channel is already closed. So we handshake here to
	// yield to the test goroutine.
	<-mc.checkAcked

	return mc.newVersionAvailable, nil
}

type mockUpdater struct {
	updateCalled bool
	callCount    int
}

func newMockUpdater() *mockUpdater {
	return new(mockUpdater)
}

func (mu *mockUpdater) Update(path string) error {
	mu.updateCalled = true
	mu.callCount++

	return nil
}

type mockCountingAsyncUpdater struct {
	closeAfterCallCount int
	errorToReturn       error
	updateCountReached  chan<- struct{}
	updateAcked         <-chan struct{}

	updateCalled bool
	callCount    int
}

func newMockCountingAsyncUpdater(closeAfterCallCount int,
	errorToReturn error,
	updateCountReached chan<- struct{},
	updateAcked <-chan struct{}) *mockCountingAsyncUpdater {
	return &mockCountingAsyncUpdater{
		errorToReturn:       errorToReturn,
		closeAfterCallCount: closeAfterCallCount,
		updateCountReached:  updateCountReached,
		updateAcked:         updateAcked,
	}
}

func (mu *mockCountingAsyncUpdater) Update(path string) error {
	mu.updateCalled = true
	mu.callCount++

	if mu.callCount == mu.closeAfterCallCount {
		// As updating will be the last step of the run() function if Update() returns an error,
		// we use Update() having been called as the signal to tell the test goroutine it is OK
		// to cancel the run() loop.
		close(mu.updateCountReached)

		// With fast poll periods, the run() goroutine will not yield the processor and hence the run()
		// loop will not have been cancelled by the test goroutine by the time Update() is called again.
		// This causes a panic as the mu.updateCountReached channel is already closed. So we handshake here to
		// yield to the test goroutine.
		<-mu.updateAcked
	}

	if mu.errorToReturn != nil {
		return mu.errorToReturn
	}

	return nil
}
