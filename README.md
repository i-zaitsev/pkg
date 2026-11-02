# pkg

Small, reusable Go packages shared across projects.

| Package | Import path | Purpose |
|---|---|---|
| `assert` | `github.com/i-zaitsev/pkg/assert` | Test assertions, plain or fluent |

```sh
go get github.com/i-zaitsev/pkg@latest
```

## assert

```go
func TestSum(t *testing.T) {
	got, err := Sum(1, 2)

	assert.NotErr(t, err)
	assert.Eq(t, got, 3)
	assert.Val(t, got).NotZero().Eq(3)
}
```

Every assertion reports through `t.Errorf`, so the test continues after a failure.

## License

[MPL-2.0](LICENSE)
