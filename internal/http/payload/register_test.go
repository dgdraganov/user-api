package payload_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dgdraganov/user-api/internal/core"
	"github.com/dgdraganov/user-api/internal/http/payload"
)

var _ = Describe("RegisterRequest", func() {
	var (
		req payload.RegisterRequest
		err error
	)
	BeforeEach(func() {
		req = payload.RegisterRequest{
			FirstName: "John",
			LastName:  "Doe",
			Age:       30,
			Email:     "john.doe@example.com",
			Password:  "securepass123",
		}
	})

	Describe("Validate", func() {

		JustBeforeEach(func() {
			err = req.Validate()
		})

		It("should succeed with valid fields", func() {
			Expect(err).To(BeNil())
		})

		DescribeTable("should fail with invalid fields", func(modify func(*payload.RegisterRequest), expectedSubstring string) {
			modify(&req)
			err = req.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedSubstring))
		},

			Entry("missing first name", func(r *payload.RegisterRequest) {
				r.FirstName = ""
			}, "first_name: cannot be blank"),

			Entry("invalid first name format", func(r *payload.RegisterRequest) {
				r.FirstName = "john"
			}, "first_name: must be in a valid format"),

			Entry("missing last name", func(r *payload.RegisterRequest) {
				r.LastName = ""
			}, "last_name: cannot be blank"),

			Entry("invalid email", func(r *payload.RegisterRequest) {
				r.Email = "not-an-email"
			}, "email: must be in a valid format"),

			Entry("too young", func(r *payload.RegisterRequest) {
				r.Age = 15
			}, "age: must be no less than 18"),

			Entry("password too short", func(r *payload.RegisterRequest) {
				r.Password = "1"
			}, "password: the length must be between 3 and 100"),
		)
	})

	Describe("ToMessage", func() {

		var (
			msg core.RegisterMessage
		)

		BeforeEach(func() {
			// req = payload.RegisterRequest{
			//  FirstName: "John",
			// 	LastName:  "Doe",
			// 	Age:       30,
			// 	Email:     "john.doe@example.com",
			// 	Password:  "securepass123",
			// }
		})

		JustBeforeEach(func() {
			msg = req.ToMessage()
		})

		It("should convert request to core.RegisterMessage", func() {
			Expect(msg.FirstName).To(Equal(req.FirstName))
			Expect(msg.LastName).To(Equal(req.LastName))
			Expect(msg.Email).To(Equal(req.Email))
			Expect(msg.Age).To(Equal(req.Age))
			Expect(msg.Password).To(Equal(req.Password))
		})
	})

	Describe("ToMap", func() {
		var m map[string]any
		JustBeforeEach(func() {
			m = req.ToMap()
		})
		It("should convert request to map correctly", func() {
			Expect(m).To(HaveKeyWithValue("first_name", "John"))
			Expect(m).To(HaveKeyWithValue("last_name", "Doe"))
			Expect(m).To(HaveKeyWithValue("email", "john.doe@example.com"))
			Expect(m).To(HaveKeyWithValue("age", 30))
		})
	})
})
