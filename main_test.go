package main

import (
	"errors"
	"testing"
	"time"
)

func TestSetupBuildsAndInstalls(t *testing.T) {
	mockObtainer := newMockObtainer(nil)
	mockTag := "mocktag"
	mockTagDeducer := newMockTagDeducer(mockTag)
	mockBuilder := newMockBuilder()
	mockPusher := newMockPusher()
	mockInstaller := newMockInstaller()
	mockSrcDirPath := ""
	mockImageName := ""
	mockInstallTimeout := time.Duration(0)

	err := setup(mockObtainer,
		mockTagDeducer,
		mockBuilder,
		mockPusher,
		mockInstaller,
		mockSrcDirPath,
		mockImageName,
		mockInstallTimeout)

	if err != nil {
		t.Errorf("expected nil error, got %q (of type %T)", err, err)
	}

	// Check the wire-up of setup() was correct
	if !mockObtainer.obtainCalled {
		t.Error("expected Obtainer.Obtain() to be called, but was not")
	}
	if !mockTagDeducer.deduceCalled {
		t.Error("expected TagDeducer.Deduce() to be called, but was not")
	}
	if !mockBuilder.buildCalled {
		t.Error("expected Builder.Build() to be called, but was not")
	}
	if !mockPusher.pushCalled {
		t.Error("expected Pusher.Push() to be called, but was not")
	}
	if !mockInstaller.installCalled {
		t.Error("expected Installer.Install() to be called, but was not")
	}
}

func TestSetupErrorsUponObtainerError(t *testing.T) {
	mockError := errors.New("mock obtainer error")
	mockObtainer := newMockObtainer(mockError)
	mockTag := "mocktag"
	mockTagDeducer := newMockTagDeducer(mockTag)
	mockBuilder := newMockBuilder()
	mockPusher := newMockPusher()
	mockInstaller := newMockInstaller()
	mockSrcDirPath := ""
	mockImageName := ""
	mockInstallTimeout := time.Duration(0)

	err := setup(mockObtainer,
		mockTagDeducer,
		mockBuilder,
		mockPusher,
		mockInstaller,
		mockSrcDirPath,
		mockImageName,
		mockInstallTimeout)

	if err == nil {
		t.Error("expected error, got nil")
	}

	t.Logf("got error %q (of type %T)", err, err)

	if !errors.Is(err, mockError) {
		t.Errorf("expected error %q, got error %q (of type %T)", mockError, err, err)
	}
}

func TestRunBuildsAndInstallsUponChange(t *testing.T) {
	mockTag := "mocktag"
	mockTagDeducer := newMockTagDeducer(mockTag)
	mockBuilder := newMockBuilder()
	mockPusher := newMockPusher()
	installed := make(chan struct{})
	installAcked := make(chan struct{})
	mockInstaller := newMockAsyncInstaller(installed, installAcked)
	mockChecker := newMockChecker(true)
	mockUpdater := newMockUpdater()
	mockSrcDirPath := ""
	mockImageName := ""
	mockInstallTimeout := time.Duration(0)
	pollPeriodDuration := time.Nanosecond
	done := make(chan struct{})

	go run(mockTagDeducer,
		mockBuilder,
		mockPusher,
		mockInstaller,
		mockChecker,
		mockUpdater,
		mockSrcDirPath,
		mockImageName,
		pollPeriodDuration,
		mockInstallTimeout,
		done)

	// Wait for the new release to be "installed", so that the done channel is not closed too early
	<-installed

	// Stop the run() goroutine from running forever
	close(done)

	// Signal to mock installer it is OK to continue
	close(installAcked)

	// Check the wire-up of run() was correct
	if !mockChecker.checkCalled {
		t.Error("expected Checker.Check() to be called, but was not")
	}
	if !mockUpdater.updateCalled {
		t.Error("expected Updater.Update() to be called, but was not")
	}
	if !mockTagDeducer.deduceCalled {
		t.Error("expected TagDeducer.Deduce() to be called, but was not")
	}
	if !mockBuilder.buildCalled {
		t.Error("expected Builder.Build() to be called, but was not")
	}
	if !mockPusher.pushCalled {
		t.Error("expected Pusher.Push() to be called, but was not")
	}
	if !mockInstaller.installCalled {
		t.Error("expected Installer.Install() to be called, but was not")
	}
}

func TestRunSkipsBuildAndInstallUponNoChange(t *testing.T) {
	mockTag := "mocktag"
	mockTagDeducer := newMockTagDeducer(mockTag)
	mockBuilder := newMockBuilder()
	mockPusher := newMockPusher()
	mockInstaller := newMockInstaller()
	checked := make(chan struct{})
	checkAcked := make(chan struct{})
	mockChecker := newMockAsyncChecker(false, checked, checkAcked)
	mockUpdater := newMockUpdater()
	mockSrcDirPath := ""
	mockImageName := ""
	mockInstallTimeout := time.Duration(0)
	pollPeriodDuration := time.Nanosecond
	done := make(chan struct{})

	go run(mockTagDeducer,
		mockBuilder,
		mockPusher,
		mockInstaller,
		mockChecker,
		mockUpdater,
		mockSrcDirPath,
		mockImageName,
		pollPeriodDuration,
		mockInstallTimeout,
		done)

	// Wait for the checker to "check", so that the done channel is not closed too early
	<-checked

	// Stop the run() goroutine from running forever
	close(done)

	// Signal to mock checker it is OK to continue
	close(checkAcked)

	// Check the wire-up of run() was correct
	if !mockChecker.checkCalled {
		t.Error("expected Checker.Check() to be called, but was not")
	}
	if mockUpdater.updateCalled {
		t.Error("expected Updater.Update() to not be called, but was")
	}
	if mockTagDeducer.deduceCalled {
		t.Error("expected TagDeducer.Deduce() to not be called, but was")
	}
	if mockBuilder.buildCalled {
		t.Error("expected Builder.Build() to not be called, but was")
	}
	if mockPusher.pushCalled {
		t.Error("expected Pusher.Push() to not be called, but was")
	}
	if mockInstaller.installCalled {
		t.Error("expected Installer.Install() to not be called, but was")
	}
}

func TestRunContinuesUponInstallError(t *testing.T) {
	callCount := 3
	mockTag := "mocktag"
	mockTagDeducer := newMockTagDeducer(mockTag)
	mockBuilder := newMockBuilder()
	mockPusher := newMockPusher()
	installCountReached := make(chan struct{})
	installAcked := make(chan struct{})
	mockError := errors.New("mock installer error")
	mockInstaller := newMockCountingAsyncInstaller(callCount, mockError, installCountReached, installAcked)
	mockChecker := newMockChecker(true)
	mockUpdater := newMockUpdater()
	mockSrcDirPath := ""
	mockImageName := ""
	mockInstallTimeout := time.Duration(0)
	pollPeriodDuration := time.Nanosecond
	done := make(chan struct{})

	go run(mockTagDeducer,
		mockBuilder,
		mockPusher,
		mockInstaller,
		mockChecker,
		mockUpdater,
		mockSrcDirPath,
		mockImageName,
		pollPeriodDuration,
		mockInstallTimeout,
		done)

	// Wait for the install attempts count to be reached, so that the done channel is not closed too early
	<-installCountReached

	// Stop the run() goroutine from running forever
	close(done)

	// Signal to mock installer it is OK to continue
	close(installAcked)

	// Check the wire-up of run() was correct
	if !mockChecker.checkCalled {
		t.Error("expected Checker.Check() to be called, but was not")
	}
	if mockChecker.callCount != callCount {
		t.Errorf("expected Checker.Check() to be called %d times, but was called %d times",
			callCount,
			mockChecker.callCount)
	}

	if !mockUpdater.updateCalled {
		t.Error("expected Updater.Update() to be called, but was not")
	}
	if mockUpdater.callCount != callCount {
		t.Errorf("expected Updater.Update() to be called %d times, but was called %d times",
			callCount,
			mockUpdater.callCount)
	}

	if !mockTagDeducer.deduceCalled {
		t.Error("expected TagDeducer.Deduce() to be called, but was not")
	}
	if mockTagDeducer.callCount != callCount {
		t.Errorf("expected TagDeducer.Deduce() to be called %d times, but was called %d times",
			callCount,
			mockTagDeducer.callCount)
	}

	if !mockBuilder.buildCalled {
		t.Error("expected Builder.Build() to be called, but was not")
	}
	if mockBuilder.callCount != callCount {
		t.Errorf("expected Builder.Build() to be called %d times, but was called %d times",
			callCount,
			mockBuilder.callCount)
	}

	if !mockPusher.pushCalled {
		t.Error("expected Pusher.Push() to be called, but was not")
	}
	if mockPusher.callCount != callCount {
		t.Errorf("expected Pusher.Push() to be called %d times, but was called %d times",
			callCount,
			mockPusher.callCount)
	}

	if !mockInstaller.installCalled {
		t.Error("expected Installer.Install() to be called, but was not")
	}
	if mockInstaller.callCount != callCount {
		t.Errorf("expected Installer.Install() to be called %d times, but was called %d times",
			callCount,
			mockInstaller.callCount)
	}
}

func TestRunContinuesUponPushError(t *testing.T) {
	callCount := 3
	mockTag := "mocktag"
	mockTagDeducer := newMockTagDeducer(mockTag)
	mockBuilder := newMockBuilder()
	mockError := errors.New("mock pusher error")
	pushCountReached := make(chan struct{})
	pushAcked := make(chan struct{})
	mockPusher := newMockCountingAsyncPusher(callCount, mockError, pushCountReached, pushAcked)
	mockInstaller := newMockInstaller()
	mockChecker := newMockChecker(true)
	mockUpdater := newMockUpdater()
	mockSrcDirPath := ""
	mockImageName := ""
	mockInstallTimeout := time.Duration(0)
	pollPeriodDuration := time.Nanosecond
	done := make(chan struct{})

	go run(mockTagDeducer,
		mockBuilder,
		mockPusher,
		mockInstaller,
		mockChecker,
		mockUpdater,
		mockSrcDirPath,
		mockImageName,
		pollPeriodDuration,
		mockInstallTimeout,
		done)

	// Wait for the push attempts count to be reached, so that the done channel is not closed too early
	<-pushCountReached

	// Stop the run() goroutine from running forever
	close(done)

	// Signal to mock pusher it is OK to continue
	close(pushAcked)

	// Check the wire-up of run() was correct
	if !mockChecker.checkCalled {
		t.Error("expected Checker.Check() to be called, but was not")
	}
	if mockChecker.callCount != callCount {
		t.Errorf("expected Checker.Check() to be called %d times, but was called %d times",
			callCount,
			mockChecker.callCount)
	}

	if !mockUpdater.updateCalled {
		t.Error("expected Updater.Update() to be called, but was not")
	}
	if mockUpdater.callCount != callCount {
		t.Errorf("expected Updater.Update() to be called %d times, but was called %d times",
			callCount,
			mockUpdater.callCount)
	}

	if !mockTagDeducer.deduceCalled {
		t.Error("expected TagDeducer.Deduce() to be called, but was not")
	}
	if mockTagDeducer.callCount != callCount {
		t.Errorf("expected TagDeducer.Deduce() to be called %d times, but was called %d times",
			callCount,
			mockTagDeducer.callCount)
	}

	if !mockBuilder.buildCalled {
		t.Error("expected Builder.Build() to be called, but was not")
	}
	if mockBuilder.callCount != callCount {
		t.Errorf("expected Builder.Build() to be called %d times, but was called %d times",
			callCount,
			mockBuilder.callCount)
	}

	if !mockPusher.pushCalled {
		t.Error("expected Pusher.Push() to be called, but was not")
	}
	if mockPusher.callCount != callCount {
		t.Errorf("expected Pusher.Push() to be called %d times, but was called %d times",
			callCount,
			mockPusher.callCount)
	}

	if mockInstaller.installCalled {
		t.Error("expected Installer.Install() to not be called, but was")
	}
}

func TestRunContinuesUponBuildError(t *testing.T) {
	callCount := 3
	mockTag := "mocktag"
	mockTagDeducer := newMockTagDeducer(mockTag)
	mockError := errors.New("mock builder error")
	buildCountReached := make(chan struct{})
	buildAcked := make(chan struct{})
	mockBuilder := newMockCountingAsyncBuilder(callCount, mockError, buildCountReached, buildAcked)
	mockPusher := newMockPusher()
	mockInstaller := newMockInstaller()
	mockChecker := newMockChecker(true)
	mockUpdater := newMockUpdater()
	mockSrcDirPath := ""
	mockImageName := ""
	mockInstallTimeout := time.Duration(0)
	pollPeriodDuration := time.Nanosecond
	done := make(chan struct{})

	go run(mockTagDeducer,
		mockBuilder,
		mockPusher,
		mockInstaller,
		mockChecker,
		mockUpdater,
		mockSrcDirPath,
		mockImageName,
		pollPeriodDuration,
		mockInstallTimeout,
		done)

	// Wait for the build attempts count to be reached, so that the done channel is not closed too early
	<-buildCountReached

	// Stop the run() goroutine from running forever
	close(done)

	// Signal to mock builder it is OK to continue
	close(buildAcked)

	// Check the wire-up of run() was correct
	if !mockChecker.checkCalled {
		t.Error("expected Checker.Check() to be called, but was not")
	}
	if mockChecker.callCount != callCount {
		t.Errorf("expected Checker.Check() to be called %d times, but was called %d times",
			callCount,
			mockChecker.callCount)
	}

	if !mockUpdater.updateCalled {
		t.Error("expected Updater.Update() to be called, but was not")
	}
	if mockUpdater.callCount != callCount {
		t.Errorf("expected Updater.Update() to be called %d times, but was called %d times",
			callCount,
			mockUpdater.callCount)
	}

	if !mockTagDeducer.deduceCalled {
		t.Error("expected TagDeducer.Deduce() to be called, but was not")
	}
	if mockTagDeducer.callCount != callCount {
		t.Errorf("expected TagDeducer.Deduce() to be called %d times, but was called %d times",
			callCount,
			mockTagDeducer.callCount)
	}

	if !mockBuilder.buildCalled {
		t.Error("expected Builder.Build() to be called, but was not")
	}
	if mockBuilder.callCount != callCount {
		t.Errorf("expected Builder.Build() to be called %d times, but was called %d times",
			callCount,
			mockBuilder.callCount)
	}

	if mockPusher.pushCalled {
		t.Error("expected Pusher.Push() to not be called, but was")
	}

	if mockInstaller.installCalled {
		t.Error("expected Installer.Install() to not be called, but was")
	}
}

func TestRunContinuesUponTagDeduceError(t *testing.T) {
	callCount := 3
	mockTag := ""
	mockError := errors.New("mock builder error")
	deduceCountReached := make(chan struct{})
	deduceAcked := make(chan struct{})
	mockTagDeducer := newMockCountingAsyncTagDeducer(callCount,
		mockTag,
		mockError,
		deduceCountReached,
		deduceAcked)
	mockBuilder := newMockBuilder()
	mockPusher := newMockPusher()
	mockInstaller := newMockInstaller()
	mockChecker := newMockChecker(true)
	mockUpdater := newMockUpdater()
	mockSrcDirPath := ""
	mockImageName := ""
	mockInstallTimeout := time.Duration(0)
	pollPeriodDuration := time.Nanosecond
	done := make(chan struct{})

	go run(mockTagDeducer,
		mockBuilder,
		mockPusher,
		mockInstaller,
		mockChecker,
		mockUpdater,
		mockSrcDirPath,
		mockImageName,
		pollPeriodDuration,
		mockInstallTimeout,
		done)

	// Wait for the tag deduce attempts count to be reached, so that the done channel is not closed too early
	<-deduceCountReached

	// Stop the run() goroutine from running forever
	close(done)

	// Signal to mock tag deducer it is OK to continue
	close(deduceAcked)

	// Check the wire-up of run() was correct
	if !mockChecker.checkCalled {
		t.Error("expected Checker.Check() to be called, but was not")
	}
	if mockChecker.callCount != callCount {
		t.Errorf("expected Checker.Check() to be called %d times, but was called %d times",
			callCount,
			mockChecker.callCount)
	}

	if !mockUpdater.updateCalled {
		t.Error("expected Updater.Update() to be called, but was not")
	}
	if mockUpdater.callCount != callCount {
		t.Errorf("expected Updater.Update() to be called %d times, but was called %d times",
			callCount,
			mockUpdater.callCount)
	}

	if !mockTagDeducer.deduceCalled {
		t.Error("expected TagDeducer.Deduce() to be called, but was not")
	}
	if mockTagDeducer.callCount != callCount {
		t.Errorf("expected TagDeducer.Deduce() to be called %d times, but was called %d times",
			callCount,
			mockTagDeducer.callCount)
	}

	if mockBuilder.buildCalled {
		t.Error("expected Builder.Build() to not be called, but was")
	}

	if mockPusher.pushCalled {
		t.Error("expected Pusher.Push() to not be called, but was")
	}

	if mockInstaller.installCalled {
		t.Error("expected Installer.Install() to not be called, but was")
	}
}

func TestRunContinuesUponUpdateError(t *testing.T) {
	callCount := 3
	mockTag := "mocktag"
	mockTagDeducer := newMockTagDeducer(mockTag)
	mockBuilder := newMockBuilder()
	mockPusher := newMockPusher()
	mockInstaller := newMockInstaller()
	mockChecker := newMockChecker(true)
	mockError := errors.New("mock builder error")
	updateCountReached := make(chan struct{})
	updateAcked := make(chan struct{})
	mockUpdater := newMockCountingAsyncUpdater(callCount, mockError, updateCountReached, updateAcked)
	mockSrcDirPath := ""
	mockImageName := ""
	mockInstallTimeout := time.Duration(0)
	pollPeriodDuration := time.Nanosecond
	done := make(chan struct{})

	go run(mockTagDeducer,
		mockBuilder,
		mockPusher,
		mockInstaller,
		mockChecker,
		mockUpdater,
		mockSrcDirPath,
		mockImageName,
		pollPeriodDuration,
		mockInstallTimeout,
		done)

	// Wait for the update attempts count to be reached, so that the done channel is not closed too early
	<-updateCountReached

	// Stop the run() goroutine from running forever
	close(done)

	// Signal to mock updater it is OK to continue
	close(updateAcked)

	// Check the wire-up of run() was correct
	if !mockChecker.checkCalled {
		t.Error("expected Checker.Check() to be called, but was not")
	}
	if mockChecker.callCount != callCount {
		t.Errorf("expected Checker.Check() to be called %d times, but was called %d times",
			callCount,
			mockChecker.callCount)
	}

	if !mockUpdater.updateCalled {
		t.Error("expected Updater.Update() to be called, but was not")
	}
	if mockUpdater.callCount != callCount {
		t.Errorf("expected Updater.Update() to be called %d times, but was called %d times",
			callCount,
			mockUpdater.callCount)
	}

	if mockTagDeducer.deduceCalled {
		t.Error("expected TagDeducer.Deduce() to not be called, but was")
	}

	if mockBuilder.buildCalled {
		t.Error("expected Builder.Build() to not be called, but was")
	}

	if mockPusher.pushCalled {
		t.Error("expected Pusher.Push() to not be called, but was")
	}

	if mockInstaller.installCalled {
		t.Error("expected Installer.Install() to not be called, but was")
	}
}

func TestRunContinuesUponCheckError(t *testing.T) {
	callCount := 3
	mockTag := "mocktag"
	mockTagDeducer := newMockTagDeducer(mockTag)
	mockBuilder := newMockBuilder()
	mockPusher := newMockPusher()
	mockInstaller := newMockInstaller()
	mockError := errors.New("mock builder error")
	checkCountReached := make(chan struct{})
	checkAcked := make(chan struct{})
	mockChecker := newMockCountingAsyncChecker(callCount, false, mockError, checkCountReached, checkAcked)
	mockUpdater := newMockUpdater()
	mockSrcDirPath := ""
	mockImageName := ""
	mockInstallTimeout := time.Duration(0)
	pollPeriodDuration := time.Nanosecond
	done := make(chan struct{})

	go run(mockTagDeducer,
		mockBuilder,
		mockPusher,
		mockInstaller,
		mockChecker,
		mockUpdater,
		mockSrcDirPath,
		mockImageName,
		pollPeriodDuration,
		mockInstallTimeout,
		done)

	// Wait for the update attempts count to be reached, so that the done channel is not closed too early
	<-checkCountReached

	// Stop the run() goroutine from running forever
	close(done)

	// Signal to mock updater it is OK to continue
	close(checkAcked)

	// Check the wire-up of run() was correct
	if !mockChecker.checkCalled {
		t.Error("expected Checker.Check() to be called, but was not")
	}
	if mockChecker.callCount != callCount {
		t.Errorf("expected Checker.Check() to be called %d times, but was called %d times",
			callCount,
			mockChecker.callCount)
	}

	if mockUpdater.updateCalled {
		t.Error("expected Updater.Update() to not be called, but was")
	}

	if mockTagDeducer.deduceCalled {
		t.Error("expected TagDeducer.Deduce() to not be called, but was")
	}

	if mockBuilder.buildCalled {
		t.Error("expected Builder.Build() to not be called, but was")
	}

	if mockPusher.pushCalled {
		t.Error("expected Pusher.Push() to not be called, but was")
	}

	if mockInstaller.installCalled {
		t.Error("expected Installer.Install() to not be called, but was")
	}
}
