package interrupt_handler_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/internal/interrupt_handler"
	"github.com/onsi/ginkgo/internal/parallel_support"
	. "github.com/onsi/ginkgo/internal/test_helpers"
	. "github.com/onsi/gomega"
)

var _ = Describe("InterruptHandler", func() {
	var interruptHandler *interrupt_handler.InterruptHandler
	Describe("Timeout interrupts", func() {
		BeforeEach(func() {
			interruptHandler = interrupt_handler.NewInterruptHandler(500*time.Millisecond, "")
		})

		AfterEach(func() {
			interruptHandler.Stop()
		})

		It("eventually closes the interrupt channel to signal an interrupt has occurred", func() {
			status := interruptHandler.Status()
			Ω(status.Interrupted).Should(BeFalse())
			Eventually(status.Channel).Should(BeClosed())

			Ω(interruptHandler.Status().Interrupted).Should(BeTrue())
		})

		It("notes the cause as 'Interrupted By Timeout'", func() {
			status := interruptHandler.Status()
			Eventually(status.Channel).Should(BeClosed())
			cause := interruptHandler.Status().Cause
			Ω(cause).Should(Equal(interrupt_handler.InterruptCauseTimeout))
			Ω(interruptHandler.InterruptMessageWithStackTraces()).Should(HavePrefix("Interrupted by Timeout\n\n"))
			Ω(interruptHandler.InterruptMessageWithStackTraces()).Should(ContainSubstring("Here's a stack trace"))
		})

		It("repeatedly triggers an interrupt every 1/10th of the registered timeout", func() {
			status := interruptHandler.Status()
			Ω(status.Interrupted).Should(BeFalse())
			Eventually(status.Channel).Should(BeClosed())

			status = interruptHandler.Status()
			Ω(status.Channel).ShouldNot(BeClosed())
			Eventually(status.Channel).Should(BeClosed())
		})
	})

	Describe("Interrupting when another Ginkgo process has aborted", func() {
		var server *parallel_support.Server
		var client parallel_support.Client
		BeforeEach(func() {
			server, client, _ = SetUpServerAndClient(2)
			interruptHandler = interrupt_handler.NewInterruptHandler(0, server.Address())
		})

		AfterEach(func() {
			server.Close()
		})

		It("interrupts when the server is told to abort", func() {
			status := interruptHandler.Status()
			Consistently(status.Channel).ShouldNot(BeClosed())
			client.PostAbort()
			Eventually(status.Channel).Should(BeClosed())
		})

		It("notes the correct cause and returns an interrupt message that does not include the stacktrace ", func() {
			status := interruptHandler.Status()
			client.PostAbort()
			Eventually(status.Channel).Should(BeClosed())
			status = interruptHandler.Status()
			Ω(status.Cause).Should(Equal(interrupt_handler.InterruptCauseAbortByOtherProcess))
			Ω(interruptHandler.InterruptMessageWithStackTraces()).Should(HavePrefix("Interrupted by Other Ginkgo Process"))
			Ω(interruptHandler.InterruptMessageWithStackTraces()).ShouldNot(ContainSubstring("Here's a stack trace"))
		})

	})
})