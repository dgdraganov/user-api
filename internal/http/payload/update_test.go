package payload_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dgdraganov/user-api/internal/http/payload"
	"github.com/dgdraganov/user-api/internal/service"
)

var _ = Describe("UpdateUserRequest", func() {

	var (
		req payload.UpdateUserRequest
		err error
	)
	BeforeEach(func() {
		req = payload.UpdateUserRequest{
			FirstName: "Alice",
			LastName:  "Smith",
			Age:       30,
			Email:     "alice@example.com",
		}
	})

	Describe("Validate", func() {
		JustBeforeEach(func() {
			err = req.Validate()
		})

		It("should succeed with valid fields", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("should fail with invalid fields",
			func(modify func(*payload.UpdateUserRequest), expectedSubstring string) {
				modify(&req)
				err = req.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedSubstring))
			},

			Entry("invalid first name format", func(r *payload.UpdateUserRequest) {
				r.FirstName = "john"
			}, "first_name: must be in a valid format"),

			Entry("invalid email", func(r *payload.UpdateUserRequest) {
				r.Email = "not-an-email"
			}, "email: must be in a valid format"),

			Entry("too young", func(r *payload.UpdateUserRequest) {
				r.Age = 14
			}, "age: must be no less than 18"),
		)
	})

	Describe("ToMessage", func() {

		var msg service.UpdateUserMessage

		JustBeforeEach(func() {
			msg = req.ToMessage()
		})

		It("should convert request to core.UpdateUserMessage", func() {
			Expect(msg.FirstName).To(Equal("Alice"))
			Expect(msg.LastName).To(Equal("Smith"))
			Expect(msg.Age).To(Equal(30))
			Expect(msg.Email).To(Equal("alice@example.com"))
		})
	})

	Describe("ToMap", func() {
		var m map[string]any

		BeforeEach(func() {
			req = payload.UpdateUserRequest{
				FirstName: "Alice",
				Age:       33,
			}
		})

		JustBeforeEach(func() {
			m = req.ToMap()
		})

		It("should include only non-zero fields", func() {
			Expect(m).To(HaveKeyWithValue("first_name", "Alice"))
			Expect(m).To(HaveKeyWithValue("age", 33))
			Expect(m).ToNot(HaveKey("last_name"))
			Expect(m).ToNot(HaveKey("email"))
		})
	})
})
