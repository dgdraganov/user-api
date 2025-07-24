package payload_test

import (
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dgdraganov/user-api/internal/http/payload"
)

var _ = Describe("UserListRequest", func() {

	var (
		req payload.UserListRequest
		err error
	)

	BeforeEach(func() {
		req = payload.UserListRequest{
			Page:     1,
			PageSize: 10,
		}
	})

	Describe("Validate", func() {

		JustBeforeEach(func() {
			err = req.Validate()
		})

		It("should succeed with valid values", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("should fail with invalid values",
			func(modify func(*payload.UserListRequest), expectedSubstring string) {
				modify(&req)
				err = req.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedSubstring))
			},

			Entry("missing page", func(r *payload.UserListRequest) {
				r.Page = -1
			}, "page: must be no less than 1"),

			Entry("missing page", func(r *payload.UserListRequest) {
				r.Page = 0
			}, "page: cannot be blank"),

			Entry("missing page_size", func(r *payload.UserListRequest) {
				r.PageSize = -1
			}, "page_size: must be no less than 1"),

			Entry("missing page_size", func(r *payload.UserListRequest) {
				r.PageSize = 0
			}, "page_size: cannot be blank"),

			Entry("integer overflow", func(r *payload.UserListRequest) {
				r.Page = int((1 << 31) - 1)
				r.PageSize = 2
			}, "would cause integer overflow"),
		)

	})

	Describe("DecodeFromURLValues", func() {
		var (
			values url.Values
		)

		JustBeforeEach(func() {
			err = req.DecodeFromURLValues(values)
		})

		When("valid parameters are provided", func() {
			BeforeEach(func() {
				values = url.Values{
					"page":      []string{"2"},
					"page_size": []string{"25"},
				}
			})

			It("should decode values correctly", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(req.Page).To(Equal(2))
				Expect(req.PageSize).To(Equal(25))
			})
		})

		DescribeTable("should fail on invalid query values",
			func(page string, pageSize string, expectedSubstring string) {
				values = url.Values{
					"page":      []string{page},
					"page_size": []string{pageSize},
				}

				err = req.DecodeFromURLValues(values)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedSubstring))
			},

			Entry("invalid page", "abc", "10", "parse page value"),
			Entry("invalid page_size", "1", "xyz", "parse page_size value"),
		)
	})
})
