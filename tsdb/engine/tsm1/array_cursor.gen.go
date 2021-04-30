// Generated by tmpl
// https://github.com/benbjohnson/tmpl
//
// DO NOT EDIT!
// Source: array_cursor.gen.go.tmpl

package tsm1

import (
	"sort"

	"github.com/influxdata/influxdb/tsdb"
)

// Array Cursors

type floatArrayAscendingCursor struct {
	cache struct {
		values Values
		pos    int
	}

	tsm struct {
		buf       *tsdb.FloatArray
		values    *tsdb.FloatArray
		pos       int
		keyCursor *KeyCursor
	}

	end int64
	res *tsdb.FloatArray
}

func newFloatArrayAscendingCursor() *floatArrayAscendingCursor {
	c := &floatArrayAscendingCursor{
		res: tsdb.NewFloatArrayLen(tsdb.DefaultMaxPointsPerBlock),
	}
	c.tsm.buf = tsdb.NewFloatArrayLen(tsdb.DefaultMaxPointsPerBlock)
	return c
}

func (c *floatArrayAscendingCursor) reset(seek, end int64, cacheValues Values, tsmKeyCursor *KeyCursor) {
	c.end = end
	c.cache.values = cacheValues
	c.cache.pos = sort.Search(len(c.cache.values), func(i int) bool {
		return c.cache.values[i].UnixNano() >= seek
	})

	c.tsm.keyCursor = tsmKeyCursor
	c.tsm.values, _ = c.tsm.keyCursor.ReadFloatArrayBlock(c.tsm.buf)
	c.tsm.pos = sort.Search(c.tsm.values.Len(), func(i int) bool {
		return c.tsm.values.Timestamps[i] >= seek
	})
}

func (c *floatArrayAscendingCursor) Err() error { return nil }

func (c *floatArrayAscendingCursor) Stats() tsdb.CursorStats {
	return tsdb.CursorStats{}
}

// close closes the cursor and any dependent cursors.
func (c *floatArrayAscendingCursor) Close() {
	if c.tsm.keyCursor != nil {
		c.tsm.keyCursor.Close()
		c.tsm.keyCursor = nil
	}
	c.cache.values = nil
	c.tsm.values = nil
}

// Next returns the next key/value for the cursor.
func (c *floatArrayAscendingCursor) Next() *tsdb.FloatArray {
	pos := 0
	cvals := c.cache.values
	tvals := c.tsm.values

	c.res.Timestamps = c.res.Timestamps[:cap(c.res.Timestamps)]
	c.res.Values = c.res.Values[:cap(c.res.Values)]

	for pos < len(c.res.Timestamps) && c.tsm.pos < len(tvals.Timestamps) && c.cache.pos < len(cvals) {
		ckey := cvals[c.cache.pos].UnixNano()
		tkey := tvals.Timestamps[c.tsm.pos]
		if ckey == tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(FloatValue).value
			c.cache.pos++
			c.tsm.pos++
		} else if ckey < tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(FloatValue).value
			c.cache.pos++
		} else {
			c.res.Timestamps[pos] = tkey
			c.res.Values[pos] = tvals.Values[c.tsm.pos]
			c.tsm.pos++
		}

		pos++

		if c.tsm.pos >= len(tvals.Timestamps) {
			tvals = c.nextTSM()
		}
	}

	if pos < len(c.res.Timestamps) {
		if c.tsm.pos < len(tvals.Timestamps) {
			if pos == 0 && len(c.res.Timestamps) >= len(tvals.Timestamps) {
				// optimization: all points can be served from TSM data because
				// we need the entire block and the block completely fits within
				// the buffer.
				copy(c.res.Timestamps, tvals.Timestamps)
				pos += copy(c.res.Values, tvals.Values)
				c.nextTSM()
			} else {
				// copy as much as we can
				n := copy(c.res.Timestamps[pos:], tvals.Timestamps[c.tsm.pos:])
				copy(c.res.Values[pos:], tvals.Values[c.tsm.pos:])
				pos += n
				c.tsm.pos += n
				if c.tsm.pos >= len(tvals.Timestamps) {
					c.nextTSM()
				}
			}
		}

		if c.cache.pos < len(cvals) {
			// TSM was exhausted
			for pos < len(c.res.Timestamps) && c.cache.pos < len(cvals) {
				c.res.Timestamps[pos] = cvals[c.cache.pos].UnixNano()
				c.res.Values[pos] = cvals[c.cache.pos].(FloatValue).value
				pos++
				c.cache.pos++
			}
		}
	}

	// Strip timestamps from after the end time.
	if pos > 0 && c.res.Timestamps[pos-1] > c.end {
		pos -= 2
		for pos >= 0 && c.res.Timestamps[pos] > c.end {
			pos--
		}
		pos++
	}

	c.res.Timestamps = c.res.Timestamps[:pos]
	c.res.Values = c.res.Values[:pos]

	return c.res
}

func (c *floatArrayAscendingCursor) nextTSM() *tsdb.FloatArray {
	c.tsm.keyCursor.Next()
	c.tsm.values, _ = c.tsm.keyCursor.ReadFloatArrayBlock(c.tsm.buf)
	c.tsm.pos = 0
	return c.tsm.values
}

type floatArrayDescendingCursor struct {
	cache struct {
		values Values
		pos    int
	}

	tsm struct {
		buf       *tsdb.FloatArray
		values    *tsdb.FloatArray
		pos       int
		keyCursor *KeyCursor
	}

	end int64
	res *tsdb.FloatArray
}

func newFloatArrayDescendingCursor() *floatArrayDescendingCursor {
	c := &floatArrayDescendingCursor{
		res: tsdb.NewFloatArrayLen(tsdb.DefaultMaxPointsPerBlock),
	}
	c.tsm.buf = tsdb.NewFloatArrayLen(tsdb.DefaultMaxPointsPerBlock)
	return c
}

func (c *floatArrayDescendingCursor) reset(seek, end int64, cacheValues Values, tsmKeyCursor *KeyCursor) {
	// Search for the time value greater than the seek time (not included)
	// and then move our position back one which will include the values in
	// our time range.
	c.end = end
	c.cache.values = cacheValues
	c.cache.pos = sort.Search(len(c.cache.values), func(i int) bool {
		return c.cache.values[i].UnixNano() > seek
	})
	c.cache.pos--

	c.tsm.keyCursor = tsmKeyCursor
	c.tsm.values, _ = c.tsm.keyCursor.ReadFloatArrayBlock(c.tsm.buf)
	c.tsm.pos = sort.Search(c.tsm.values.Len(), func(i int) bool {
		return c.tsm.values.Timestamps[i] > seek
	})
	c.tsm.pos--
}

func (c *floatArrayDescendingCursor) Err() error { return nil }

func (c *floatArrayDescendingCursor) Stats() tsdb.CursorStats {
	return tsdb.CursorStats{}
}

func (c *floatArrayDescendingCursor) Close() {
	if c.tsm.keyCursor != nil {
		c.tsm.keyCursor.Close()
		c.tsm.keyCursor = nil
	}
	c.cache.values = nil
	c.tsm.values = nil
}

func (c *floatArrayDescendingCursor) Next() *tsdb.FloatArray {
	pos := 0
	cvals := c.cache.values
	tvals := c.tsm.values

	c.res.Timestamps = c.res.Timestamps[:cap(c.res.Timestamps)]
	c.res.Values = c.res.Values[:cap(c.res.Values)]

	for pos < len(c.res.Timestamps) && c.tsm.pos >= 0 && c.cache.pos >= 0 {
		ckey := cvals[c.cache.pos].UnixNano()
		tkey := tvals.Timestamps[c.tsm.pos]
		if ckey == tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(FloatValue).value
			c.cache.pos--
			c.tsm.pos--
		} else if ckey > tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(FloatValue).value
			c.cache.pos--
		} else {
			c.res.Timestamps[pos] = tkey
			c.res.Values[pos] = tvals.Values[c.tsm.pos]
			c.tsm.pos--
		}

		pos++

		if c.tsm.pos < 0 {
			tvals = c.nextTSM()
		}
	}

	if pos < len(c.res.Timestamps) {
		// cache was exhausted
		if c.tsm.pos >= 0 {
			for pos < len(c.res.Timestamps) && c.tsm.pos >= 0 {
				c.res.Timestamps[pos] = tvals.Timestamps[c.tsm.pos]
				c.res.Values[pos] = tvals.Values[c.tsm.pos]
				pos++
				c.tsm.pos--
				if c.tsm.pos < 0 {
					tvals = c.nextTSM()
				}
			}
		}

		if c.cache.pos >= 0 {
			// TSM was exhausted
			for pos < len(c.res.Timestamps) && c.cache.pos >= 0 {
				c.res.Timestamps[pos] = cvals[c.cache.pos].UnixNano()
				c.res.Values[pos] = cvals[c.cache.pos].(FloatValue).value
				pos++
				c.cache.pos--
			}
		}
	}

	// Strip timestamps from before the end time.
	if pos > 0 && c.res.Timestamps[pos-1] < c.end {
		pos -= 2
		for pos >= 0 && c.res.Timestamps[pos] < c.end {
			pos--
		}
		pos++
	}

	c.res.Timestamps = c.res.Timestamps[:pos]
	c.res.Values = c.res.Values[:pos]

	return c.res
}

func (c *floatArrayDescendingCursor) nextTSM() *tsdb.FloatArray {
	c.tsm.keyCursor.Next()
	c.tsm.values, _ = c.tsm.keyCursor.ReadFloatArrayBlock(c.tsm.buf)
	c.tsm.pos = len(c.tsm.values.Timestamps) - 1
	return c.tsm.values
}

type integerArrayAscendingCursor struct {
	cache struct {
		values Values
		pos    int
	}

	tsm struct {
		buf       *tsdb.IntegerArray
		values    *tsdb.IntegerArray
		pos       int
		keyCursor *KeyCursor
	}

	end int64
	res *tsdb.IntegerArray
}

func newIntegerArrayAscendingCursor() *integerArrayAscendingCursor {
	c := &integerArrayAscendingCursor{
		res: tsdb.NewIntegerArrayLen(tsdb.DefaultMaxPointsPerBlock),
	}
	c.tsm.buf = tsdb.NewIntegerArrayLen(tsdb.DefaultMaxPointsPerBlock)
	return c
}

func (c *integerArrayAscendingCursor) reset(seek, end int64, cacheValues Values, tsmKeyCursor *KeyCursor) {
	c.end = end
	c.cache.values = cacheValues
	c.cache.pos = sort.Search(len(c.cache.values), func(i int) bool {
		return c.cache.values[i].UnixNano() >= seek
	})

	c.tsm.keyCursor = tsmKeyCursor
	c.tsm.values, _ = c.tsm.keyCursor.ReadIntegerArrayBlock(c.tsm.buf)
	c.tsm.pos = sort.Search(c.tsm.values.Len(), func(i int) bool {
		return c.tsm.values.Timestamps[i] >= seek
	})
}

func (c *integerArrayAscendingCursor) Err() error { return nil }

func (c *integerArrayAscendingCursor) Stats() tsdb.CursorStats {
	return tsdb.CursorStats{}
}

// close closes the cursor and any dependent cursors.
func (c *integerArrayAscendingCursor) Close() {
	if c.tsm.keyCursor != nil {
		c.tsm.keyCursor.Close()
		c.tsm.keyCursor = nil
	}
	c.cache.values = nil
	c.tsm.values = nil
}

// Next returns the next key/value for the cursor.
func (c *integerArrayAscendingCursor) Next() *tsdb.IntegerArray {
	pos := 0
	cvals := c.cache.values
	tvals := c.tsm.values

	c.res.Timestamps = c.res.Timestamps[:cap(c.res.Timestamps)]
	c.res.Values = c.res.Values[:cap(c.res.Values)]

	for pos < len(c.res.Timestamps) && c.tsm.pos < len(tvals.Timestamps) && c.cache.pos < len(cvals) {
		ckey := cvals[c.cache.pos].UnixNano()
		tkey := tvals.Timestamps[c.tsm.pos]
		if ckey == tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(IntegerValue).value
			c.cache.pos++
			c.tsm.pos++
		} else if ckey < tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(IntegerValue).value
			c.cache.pos++
		} else {
			c.res.Timestamps[pos] = tkey
			c.res.Values[pos] = tvals.Values[c.tsm.pos]
			c.tsm.pos++
		}

		pos++

		if c.tsm.pos >= len(tvals.Timestamps) {
			tvals = c.nextTSM()
		}
	}

	if pos < len(c.res.Timestamps) {
		if c.tsm.pos < len(tvals.Timestamps) {
			if pos == 0 && len(c.res.Timestamps) >= len(tvals.Timestamps) {
				// optimization: all points can be served from TSM data because
				// we need the entire block and the block completely fits within
				// the buffer.
				copy(c.res.Timestamps, tvals.Timestamps)
				pos += copy(c.res.Values, tvals.Values)
				c.nextTSM()
			} else {
				// copy as much as we can
				n := copy(c.res.Timestamps[pos:], tvals.Timestamps[c.tsm.pos:])
				copy(c.res.Values[pos:], tvals.Values[c.tsm.pos:])
				pos += n
				c.tsm.pos += n
				if c.tsm.pos >= len(tvals.Timestamps) {
					c.nextTSM()
				}
			}
		}

		if c.cache.pos < len(cvals) {
			// TSM was exhausted
			for pos < len(c.res.Timestamps) && c.cache.pos < len(cvals) {
				c.res.Timestamps[pos] = cvals[c.cache.pos].UnixNano()
				c.res.Values[pos] = cvals[c.cache.pos].(IntegerValue).value
				pos++
				c.cache.pos++
			}
		}
	}

	// Strip timestamps from after the end time.
	if pos > 0 && c.res.Timestamps[pos-1] > c.end {
		pos -= 2
		for pos >= 0 && c.res.Timestamps[pos] > c.end {
			pos--
		}
		pos++
	}

	c.res.Timestamps = c.res.Timestamps[:pos]
	c.res.Values = c.res.Values[:pos]

	return c.res
}

func (c *integerArrayAscendingCursor) nextTSM() *tsdb.IntegerArray {
	c.tsm.keyCursor.Next()
	c.tsm.values, _ = c.tsm.keyCursor.ReadIntegerArrayBlock(c.tsm.buf)
	c.tsm.pos = 0
	return c.tsm.values
}

type integerArrayDescendingCursor struct {
	cache struct {
		values Values
		pos    int
	}

	tsm struct {
		buf       *tsdb.IntegerArray
		values    *tsdb.IntegerArray
		pos       int
		keyCursor *KeyCursor
	}

	end int64
	res *tsdb.IntegerArray
}

func newIntegerArrayDescendingCursor() *integerArrayDescendingCursor {
	c := &integerArrayDescendingCursor{
		res: tsdb.NewIntegerArrayLen(tsdb.DefaultMaxPointsPerBlock),
	}
	c.tsm.buf = tsdb.NewIntegerArrayLen(tsdb.DefaultMaxPointsPerBlock)
	return c
}

func (c *integerArrayDescendingCursor) reset(seek, end int64, cacheValues Values, tsmKeyCursor *KeyCursor) {
	// Search for the time value greater than the seek time (not included)
	// and then move our position back one which will include the values in
	// our time range.
	c.end = end
	c.cache.values = cacheValues
	c.cache.pos = sort.Search(len(c.cache.values), func(i int) bool {
		return c.cache.values[i].UnixNano() > seek
	})
	c.cache.pos--

	c.tsm.keyCursor = tsmKeyCursor
	c.tsm.values, _ = c.tsm.keyCursor.ReadIntegerArrayBlock(c.tsm.buf)
	c.tsm.pos = sort.Search(c.tsm.values.Len(), func(i int) bool {
		return c.tsm.values.Timestamps[i] > seek
	})
	c.tsm.pos--
}

func (c *integerArrayDescendingCursor) Err() error { return nil }

func (c *integerArrayDescendingCursor) Stats() tsdb.CursorStats {
	return tsdb.CursorStats{}
}

func (c *integerArrayDescendingCursor) Close() {
	if c.tsm.keyCursor != nil {
		c.tsm.keyCursor.Close()
		c.tsm.keyCursor = nil
	}
	c.cache.values = nil
	c.tsm.values = nil
}

func (c *integerArrayDescendingCursor) Next() *tsdb.IntegerArray {
	pos := 0
	cvals := c.cache.values
	tvals := c.tsm.values

	c.res.Timestamps = c.res.Timestamps[:cap(c.res.Timestamps)]
	c.res.Values = c.res.Values[:cap(c.res.Values)]

	for pos < len(c.res.Timestamps) && c.tsm.pos >= 0 && c.cache.pos >= 0 {
		ckey := cvals[c.cache.pos].UnixNano()
		tkey := tvals.Timestamps[c.tsm.pos]
		if ckey == tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(IntegerValue).value
			c.cache.pos--
			c.tsm.pos--
		} else if ckey > tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(IntegerValue).value
			c.cache.pos--
		} else {
			c.res.Timestamps[pos] = tkey
			c.res.Values[pos] = tvals.Values[c.tsm.pos]
			c.tsm.pos--
		}

		pos++

		if c.tsm.pos < 0 {
			tvals = c.nextTSM()
		}
	}

	if pos < len(c.res.Timestamps) {
		// cache was exhausted
		if c.tsm.pos >= 0 {
			for pos < len(c.res.Timestamps) && c.tsm.pos >= 0 {
				c.res.Timestamps[pos] = tvals.Timestamps[c.tsm.pos]
				c.res.Values[pos] = tvals.Values[c.tsm.pos]
				pos++
				c.tsm.pos--
				if c.tsm.pos < 0 {
					tvals = c.nextTSM()
				}
			}
		}

		if c.cache.pos >= 0 {
			// TSM was exhausted
			for pos < len(c.res.Timestamps) && c.cache.pos >= 0 {
				c.res.Timestamps[pos] = cvals[c.cache.pos].UnixNano()
				c.res.Values[pos] = cvals[c.cache.pos].(IntegerValue).value
				pos++
				c.cache.pos--
			}
		}
	}

	// Strip timestamps from before the end time.
	if pos > 0 && c.res.Timestamps[pos-1] < c.end {
		pos -= 2
		for pos >= 0 && c.res.Timestamps[pos] < c.end {
			pos--
		}
		pos++
	}

	c.res.Timestamps = c.res.Timestamps[:pos]
	c.res.Values = c.res.Values[:pos]

	return c.res
}

func (c *integerArrayDescendingCursor) nextTSM() *tsdb.IntegerArray {
	c.tsm.keyCursor.Next()
	c.tsm.values, _ = c.tsm.keyCursor.ReadIntegerArrayBlock(c.tsm.buf)
	c.tsm.pos = len(c.tsm.values.Timestamps) - 1
	return c.tsm.values
}

type unsignedArrayAscendingCursor struct {
	cache struct {
		values Values
		pos    int
	}

	tsm struct {
		buf       *tsdb.UnsignedArray
		values    *tsdb.UnsignedArray
		pos       int
		keyCursor *KeyCursor
	}

	end int64
	res *tsdb.UnsignedArray
}

func newUnsignedArrayAscendingCursor() *unsignedArrayAscendingCursor {
	c := &unsignedArrayAscendingCursor{
		res: tsdb.NewUnsignedArrayLen(tsdb.DefaultMaxPointsPerBlock),
	}
	c.tsm.buf = tsdb.NewUnsignedArrayLen(tsdb.DefaultMaxPointsPerBlock)
	return c
}

func (c *unsignedArrayAscendingCursor) reset(seek, end int64, cacheValues Values, tsmKeyCursor *KeyCursor) {
	c.end = end
	c.cache.values = cacheValues
	c.cache.pos = sort.Search(len(c.cache.values), func(i int) bool {
		return c.cache.values[i].UnixNano() >= seek
	})

	c.tsm.keyCursor = tsmKeyCursor
	c.tsm.values, _ = c.tsm.keyCursor.ReadUnsignedArrayBlock(c.tsm.buf)
	c.tsm.pos = sort.Search(c.tsm.values.Len(), func(i int) bool {
		return c.tsm.values.Timestamps[i] >= seek
	})
}

func (c *unsignedArrayAscendingCursor) Err() error { return nil }

func (c *unsignedArrayAscendingCursor) Stats() tsdb.CursorStats {
	return tsdb.CursorStats{}
}

// close closes the cursor and any dependent cursors.
func (c *unsignedArrayAscendingCursor) Close() {
	if c.tsm.keyCursor != nil {
		c.tsm.keyCursor.Close()
		c.tsm.keyCursor = nil
	}
	c.cache.values = nil
	c.tsm.values = nil
}

// Next returns the next key/value for the cursor.
func (c *unsignedArrayAscendingCursor) Next() *tsdb.UnsignedArray {
	pos := 0
	cvals := c.cache.values
	tvals := c.tsm.values

	c.res.Timestamps = c.res.Timestamps[:cap(c.res.Timestamps)]
	c.res.Values = c.res.Values[:cap(c.res.Values)]

	for pos < len(c.res.Timestamps) && c.tsm.pos < len(tvals.Timestamps) && c.cache.pos < len(cvals) {
		ckey := cvals[c.cache.pos].UnixNano()
		tkey := tvals.Timestamps[c.tsm.pos]
		if ckey == tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(UnsignedValue).value
			c.cache.pos++
			c.tsm.pos++
		} else if ckey < tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(UnsignedValue).value
			c.cache.pos++
		} else {
			c.res.Timestamps[pos] = tkey
			c.res.Values[pos] = tvals.Values[c.tsm.pos]
			c.tsm.pos++
		}

		pos++

		if c.tsm.pos >= len(tvals.Timestamps) {
			tvals = c.nextTSM()
		}
	}

	if pos < len(c.res.Timestamps) {
		if c.tsm.pos < len(tvals.Timestamps) {
			if pos == 0 && len(c.res.Timestamps) >= len(tvals.Timestamps) {
				// optimization: all points can be served from TSM data because
				// we need the entire block and the block completely fits within
				// the buffer.
				copy(c.res.Timestamps, tvals.Timestamps)
				pos += copy(c.res.Values, tvals.Values)
				c.nextTSM()
			} else {
				// copy as much as we can
				n := copy(c.res.Timestamps[pos:], tvals.Timestamps[c.tsm.pos:])
				copy(c.res.Values[pos:], tvals.Values[c.tsm.pos:])
				pos += n
				c.tsm.pos += n
				if c.tsm.pos >= len(tvals.Timestamps) {
					c.nextTSM()
				}
			}
		}

		if c.cache.pos < len(cvals) {
			// TSM was exhausted
			for pos < len(c.res.Timestamps) && c.cache.pos < len(cvals) {
				c.res.Timestamps[pos] = cvals[c.cache.pos].UnixNano()
				c.res.Values[pos] = cvals[c.cache.pos].(UnsignedValue).value
				pos++
				c.cache.pos++
			}
		}
	}

	// Strip timestamps from after the end time.
	if pos > 0 && c.res.Timestamps[pos-1] > c.end {
		pos -= 2
		for pos >= 0 && c.res.Timestamps[pos] > c.end {
			pos--
		}
		pos++
	}

	c.res.Timestamps = c.res.Timestamps[:pos]
	c.res.Values = c.res.Values[:pos]

	return c.res
}

func (c *unsignedArrayAscendingCursor) nextTSM() *tsdb.UnsignedArray {
	c.tsm.keyCursor.Next()
	c.tsm.values, _ = c.tsm.keyCursor.ReadUnsignedArrayBlock(c.tsm.buf)
	c.tsm.pos = 0
	return c.tsm.values
}

type unsignedArrayDescendingCursor struct {
	cache struct {
		values Values
		pos    int
	}

	tsm struct {
		buf       *tsdb.UnsignedArray
		values    *tsdb.UnsignedArray
		pos       int
		keyCursor *KeyCursor
	}

	end int64
	res *tsdb.UnsignedArray
}

func newUnsignedArrayDescendingCursor() *unsignedArrayDescendingCursor {
	c := &unsignedArrayDescendingCursor{
		res: tsdb.NewUnsignedArrayLen(tsdb.DefaultMaxPointsPerBlock),
	}
	c.tsm.buf = tsdb.NewUnsignedArrayLen(tsdb.DefaultMaxPointsPerBlock)
	return c
}

func (c *unsignedArrayDescendingCursor) reset(seek, end int64, cacheValues Values, tsmKeyCursor *KeyCursor) {
	// Search for the time value greater than the seek time (not included)
	// and then move our position back one which will include the values in
	// our time range.
	c.end = end
	c.cache.values = cacheValues
	c.cache.pos = sort.Search(len(c.cache.values), func(i int) bool {
		return c.cache.values[i].UnixNano() > seek
	})
	c.cache.pos--

	c.tsm.keyCursor = tsmKeyCursor
	c.tsm.values, _ = c.tsm.keyCursor.ReadUnsignedArrayBlock(c.tsm.buf)
	c.tsm.pos = sort.Search(c.tsm.values.Len(), func(i int) bool {
		return c.tsm.values.Timestamps[i] > seek
	})
	c.tsm.pos--
}

func (c *unsignedArrayDescendingCursor) Err() error { return nil }

func (c *unsignedArrayDescendingCursor) Stats() tsdb.CursorStats {
	return tsdb.CursorStats{}
}

func (c *unsignedArrayDescendingCursor) Close() {
	if c.tsm.keyCursor != nil {
		c.tsm.keyCursor.Close()
		c.tsm.keyCursor = nil
	}
	c.cache.values = nil
	c.tsm.values = nil
}

func (c *unsignedArrayDescendingCursor) Next() *tsdb.UnsignedArray {
	pos := 0
	cvals := c.cache.values
	tvals := c.tsm.values

	c.res.Timestamps = c.res.Timestamps[:cap(c.res.Timestamps)]
	c.res.Values = c.res.Values[:cap(c.res.Values)]

	for pos < len(c.res.Timestamps) && c.tsm.pos >= 0 && c.cache.pos >= 0 {
		ckey := cvals[c.cache.pos].UnixNano()
		tkey := tvals.Timestamps[c.tsm.pos]
		if ckey == tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(UnsignedValue).value
			c.cache.pos--
			c.tsm.pos--
		} else if ckey > tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(UnsignedValue).value
			c.cache.pos--
		} else {
			c.res.Timestamps[pos] = tkey
			c.res.Values[pos] = tvals.Values[c.tsm.pos]
			c.tsm.pos--
		}

		pos++

		if c.tsm.pos < 0 {
			tvals = c.nextTSM()
		}
	}

	if pos < len(c.res.Timestamps) {
		// cache was exhausted
		if c.tsm.pos >= 0 {
			for pos < len(c.res.Timestamps) && c.tsm.pos >= 0 {
				c.res.Timestamps[pos] = tvals.Timestamps[c.tsm.pos]
				c.res.Values[pos] = tvals.Values[c.tsm.pos]
				pos++
				c.tsm.pos--
				if c.tsm.pos < 0 {
					tvals = c.nextTSM()
				}
			}
		}

		if c.cache.pos >= 0 {
			// TSM was exhausted
			for pos < len(c.res.Timestamps) && c.cache.pos >= 0 {
				c.res.Timestamps[pos] = cvals[c.cache.pos].UnixNano()
				c.res.Values[pos] = cvals[c.cache.pos].(UnsignedValue).value
				pos++
				c.cache.pos--
			}
		}
	}

	// Strip timestamps from before the end time.
	if pos > 0 && c.res.Timestamps[pos-1] < c.end {
		pos -= 2
		for pos >= 0 && c.res.Timestamps[pos] < c.end {
			pos--
		}
		pos++
	}

	c.res.Timestamps = c.res.Timestamps[:pos]
	c.res.Values = c.res.Values[:pos]

	return c.res
}

func (c *unsignedArrayDescendingCursor) nextTSM() *tsdb.UnsignedArray {
	c.tsm.keyCursor.Next()
	c.tsm.values, _ = c.tsm.keyCursor.ReadUnsignedArrayBlock(c.tsm.buf)
	c.tsm.pos = len(c.tsm.values.Timestamps) - 1
	return c.tsm.values
}

type stringArrayAscendingCursor struct {
	cache struct {
		values Values
		pos    int
	}

	tsm struct {
		buf       *tsdb.StringArray
		values    *tsdb.StringArray
		pos       int
		keyCursor *KeyCursor
	}

	end int64
	res *tsdb.StringArray
}

func newStringArrayAscendingCursor() *stringArrayAscendingCursor {
	c := &stringArrayAscendingCursor{
		res: tsdb.NewStringArrayLen(tsdb.DefaultMaxPointsPerBlock),
	}
	c.tsm.buf = tsdb.NewStringArrayLen(tsdb.DefaultMaxPointsPerBlock)
	return c
}

func (c *stringArrayAscendingCursor) reset(seek, end int64, cacheValues Values, tsmKeyCursor *KeyCursor) {
	c.end = end
	c.cache.values = cacheValues
	c.cache.pos = sort.Search(len(c.cache.values), func(i int) bool {
		return c.cache.values[i].UnixNano() >= seek
	})

	c.tsm.keyCursor = tsmKeyCursor
	c.tsm.values, _ = c.tsm.keyCursor.ReadStringArrayBlock(c.tsm.buf)
	c.tsm.pos = sort.Search(c.tsm.values.Len(), func(i int) bool {
		return c.tsm.values.Timestamps[i] >= seek
	})
}

func (c *stringArrayAscendingCursor) Err() error { return nil }

func (c *stringArrayAscendingCursor) Stats() tsdb.CursorStats {
	return tsdb.CursorStats{}
}

// close closes the cursor and any dependent cursors.
func (c *stringArrayAscendingCursor) Close() {
	if c.tsm.keyCursor != nil {
		c.tsm.keyCursor.Close()
		c.tsm.keyCursor = nil
	}
	c.cache.values = nil
	c.tsm.values = nil
}

// Next returns the next key/value for the cursor.
func (c *stringArrayAscendingCursor) Next() *tsdb.StringArray {
	pos := 0
	cvals := c.cache.values
	tvals := c.tsm.values

	c.res.Timestamps = c.res.Timestamps[:cap(c.res.Timestamps)]
	c.res.Values = c.res.Values[:cap(c.res.Values)]

	for pos < len(c.res.Timestamps) && c.tsm.pos < len(tvals.Timestamps) && c.cache.pos < len(cvals) {
		ckey := cvals[c.cache.pos].UnixNano()
		tkey := tvals.Timestamps[c.tsm.pos]
		if ckey == tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(StringValue).value
			c.cache.pos++
			c.tsm.pos++
		} else if ckey < tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(StringValue).value
			c.cache.pos++
		} else {
			c.res.Timestamps[pos] = tkey
			c.res.Values[pos] = tvals.Values[c.tsm.pos]
			c.tsm.pos++
		}

		pos++

		if c.tsm.pos >= len(tvals.Timestamps) {
			tvals = c.nextTSM()
		}
	}

	if pos < len(c.res.Timestamps) {
		if c.tsm.pos < len(tvals.Timestamps) {
			if pos == 0 && len(c.res.Timestamps) >= len(tvals.Timestamps) {
				// optimization: all points can be served from TSM data because
				// we need the entire block and the block completely fits within
				// the buffer.
				copy(c.res.Timestamps, tvals.Timestamps)
				pos += copy(c.res.Values, tvals.Values)
				c.nextTSM()
			} else {
				// copy as much as we can
				n := copy(c.res.Timestamps[pos:], tvals.Timestamps[c.tsm.pos:])
				copy(c.res.Values[pos:], tvals.Values[c.tsm.pos:])
				pos += n
				c.tsm.pos += n
				if c.tsm.pos >= len(tvals.Timestamps) {
					c.nextTSM()
				}
			}
		}

		if c.cache.pos < len(cvals) {
			// TSM was exhausted
			for pos < len(c.res.Timestamps) && c.cache.pos < len(cvals) {
				c.res.Timestamps[pos] = cvals[c.cache.pos].UnixNano()
				c.res.Values[pos] = cvals[c.cache.pos].(StringValue).value
				pos++
				c.cache.pos++
			}
		}
	}

	// Strip timestamps from after the end time.
	if pos > 0 && c.res.Timestamps[pos-1] > c.end {
		pos -= 2
		for pos >= 0 && c.res.Timestamps[pos] > c.end {
			pos--
		}
		pos++
	}

	c.res.Timestamps = c.res.Timestamps[:pos]
	c.res.Values = c.res.Values[:pos]

	return c.res
}

func (c *stringArrayAscendingCursor) nextTSM() *tsdb.StringArray {
	c.tsm.keyCursor.Next()
	c.tsm.values, _ = c.tsm.keyCursor.ReadStringArrayBlock(c.tsm.buf)
	c.tsm.pos = 0
	return c.tsm.values
}

type stringArrayDescendingCursor struct {
	cache struct {
		values Values
		pos    int
	}

	tsm struct {
		buf       *tsdb.StringArray
		values    *tsdb.StringArray
		pos       int
		keyCursor *KeyCursor
	}

	end int64
	res *tsdb.StringArray
}

func newStringArrayDescendingCursor() *stringArrayDescendingCursor {
	c := &stringArrayDescendingCursor{
		res: tsdb.NewStringArrayLen(tsdb.DefaultMaxPointsPerBlock),
	}
	c.tsm.buf = tsdb.NewStringArrayLen(tsdb.DefaultMaxPointsPerBlock)
	return c
}

func (c *stringArrayDescendingCursor) reset(seek, end int64, cacheValues Values, tsmKeyCursor *KeyCursor) {
	// Search for the time value greater than the seek time (not included)
	// and then move our position back one which will include the values in
	// our time range.
	c.end = end
	c.cache.values = cacheValues
	c.cache.pos = sort.Search(len(c.cache.values), func(i int) bool {
		return c.cache.values[i].UnixNano() > seek
	})
	c.cache.pos--

	c.tsm.keyCursor = tsmKeyCursor
	c.tsm.values, _ = c.tsm.keyCursor.ReadStringArrayBlock(c.tsm.buf)
	c.tsm.pos = sort.Search(c.tsm.values.Len(), func(i int) bool {
		return c.tsm.values.Timestamps[i] > seek
	})
	c.tsm.pos--
}

func (c *stringArrayDescendingCursor) Err() error { return nil }

func (c *stringArrayDescendingCursor) Stats() tsdb.CursorStats {
	return tsdb.CursorStats{}
}

func (c *stringArrayDescendingCursor) Close() {
	if c.tsm.keyCursor != nil {
		c.tsm.keyCursor.Close()
		c.tsm.keyCursor = nil
	}
	c.cache.values = nil
	c.tsm.values = nil
}

func (c *stringArrayDescendingCursor) Next() *tsdb.StringArray {
	pos := 0
	cvals := c.cache.values
	tvals := c.tsm.values

	c.res.Timestamps = c.res.Timestamps[:cap(c.res.Timestamps)]
	c.res.Values = c.res.Values[:cap(c.res.Values)]

	for pos < len(c.res.Timestamps) && c.tsm.pos >= 0 && c.cache.pos >= 0 {
		ckey := cvals[c.cache.pos].UnixNano()
		tkey := tvals.Timestamps[c.tsm.pos]
		if ckey == tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(StringValue).value
			c.cache.pos--
			c.tsm.pos--
		} else if ckey > tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(StringValue).value
			c.cache.pos--
		} else {
			c.res.Timestamps[pos] = tkey
			c.res.Values[pos] = tvals.Values[c.tsm.pos]
			c.tsm.pos--
		}

		pos++

		if c.tsm.pos < 0 {
			tvals = c.nextTSM()
		}
	}

	if pos < len(c.res.Timestamps) {
		// cache was exhausted
		if c.tsm.pos >= 0 {
			for pos < len(c.res.Timestamps) && c.tsm.pos >= 0 {
				c.res.Timestamps[pos] = tvals.Timestamps[c.tsm.pos]
				c.res.Values[pos] = tvals.Values[c.tsm.pos]
				pos++
				c.tsm.pos--
				if c.tsm.pos < 0 {
					tvals = c.nextTSM()
				}
			}
		}

		if c.cache.pos >= 0 {
			// TSM was exhausted
			for pos < len(c.res.Timestamps) && c.cache.pos >= 0 {
				c.res.Timestamps[pos] = cvals[c.cache.pos].UnixNano()
				c.res.Values[pos] = cvals[c.cache.pos].(StringValue).value
				pos++
				c.cache.pos--
			}
		}
	}

	// Strip timestamps from before the end time.
	if pos > 0 && c.res.Timestamps[pos-1] < c.end {
		pos -= 2
		for pos >= 0 && c.res.Timestamps[pos] < c.end {
			pos--
		}
		pos++
	}

	c.res.Timestamps = c.res.Timestamps[:pos]
	c.res.Values = c.res.Values[:pos]

	return c.res
}

func (c *stringArrayDescendingCursor) nextTSM() *tsdb.StringArray {
	c.tsm.keyCursor.Next()
	c.tsm.values, _ = c.tsm.keyCursor.ReadStringArrayBlock(c.tsm.buf)
	c.tsm.pos = len(c.tsm.values.Timestamps) - 1
	return c.tsm.values
}

type booleanArrayAscendingCursor struct {
	cache struct {
		values Values
		pos    int
	}

	tsm struct {
		buf       *tsdb.BooleanArray
		values    *tsdb.BooleanArray
		pos       int
		keyCursor *KeyCursor
	}

	end int64
	res *tsdb.BooleanArray
}

func newBooleanArrayAscendingCursor() *booleanArrayAscendingCursor {
	c := &booleanArrayAscendingCursor{
		res: tsdb.NewBooleanArrayLen(tsdb.DefaultMaxPointsPerBlock),
	}
	c.tsm.buf = tsdb.NewBooleanArrayLen(tsdb.DefaultMaxPointsPerBlock)
	return c
}

func (c *booleanArrayAscendingCursor) reset(seek, end int64, cacheValues Values, tsmKeyCursor *KeyCursor) {
	c.end = end
	c.cache.values = cacheValues
	c.cache.pos = sort.Search(len(c.cache.values), func(i int) bool {
		return c.cache.values[i].UnixNano() >= seek
	})

	c.tsm.keyCursor = tsmKeyCursor
	c.tsm.values, _ = c.tsm.keyCursor.ReadBooleanArrayBlock(c.tsm.buf)
	c.tsm.pos = sort.Search(c.tsm.values.Len(), func(i int) bool {
		return c.tsm.values.Timestamps[i] >= seek
	})
}

func (c *booleanArrayAscendingCursor) Err() error { return nil }

func (c *booleanArrayAscendingCursor) Stats() tsdb.CursorStats {
	return tsdb.CursorStats{}
}

// close closes the cursor and any dependent cursors.
func (c *booleanArrayAscendingCursor) Close() {
	if c.tsm.keyCursor != nil {
		c.tsm.keyCursor.Close()
		c.tsm.keyCursor = nil
	}
	c.cache.values = nil
	c.tsm.values = nil
}

// Next returns the next key/value for the cursor.
func (c *booleanArrayAscendingCursor) Next() *tsdb.BooleanArray {
	pos := 0
	cvals := c.cache.values
	tvals := c.tsm.values

	c.res.Timestamps = c.res.Timestamps[:cap(c.res.Timestamps)]
	c.res.Values = c.res.Values[:cap(c.res.Values)]

	for pos < len(c.res.Timestamps) && c.tsm.pos < len(tvals.Timestamps) && c.cache.pos < len(cvals) {
		ckey := cvals[c.cache.pos].UnixNano()
		tkey := tvals.Timestamps[c.tsm.pos]
		if ckey == tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(BooleanValue).value
			c.cache.pos++
			c.tsm.pos++
		} else if ckey < tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(BooleanValue).value
			c.cache.pos++
		} else {
			c.res.Timestamps[pos] = tkey
			c.res.Values[pos] = tvals.Values[c.tsm.pos]
			c.tsm.pos++
		}

		pos++

		if c.tsm.pos >= len(tvals.Timestamps) {
			tvals = c.nextTSM()
		}
	}

	if pos < len(c.res.Timestamps) {
		if c.tsm.pos < len(tvals.Timestamps) {
			if pos == 0 && len(c.res.Timestamps) >= len(tvals.Timestamps) {
				// optimization: all points can be served from TSM data because
				// we need the entire block and the block completely fits within
				// the buffer.
				copy(c.res.Timestamps, tvals.Timestamps)
				pos += copy(c.res.Values, tvals.Values)
				c.nextTSM()
			} else {
				// copy as much as we can
				n := copy(c.res.Timestamps[pos:], tvals.Timestamps[c.tsm.pos:])
				copy(c.res.Values[pos:], tvals.Values[c.tsm.pos:])
				pos += n
				c.tsm.pos += n
				if c.tsm.pos >= len(tvals.Timestamps) {
					c.nextTSM()
				}
			}
		}

		if c.cache.pos < len(cvals) {
			// TSM was exhausted
			for pos < len(c.res.Timestamps) && c.cache.pos < len(cvals) {
				c.res.Timestamps[pos] = cvals[c.cache.pos].UnixNano()
				c.res.Values[pos] = cvals[c.cache.pos].(BooleanValue).value
				pos++
				c.cache.pos++
			}
		}
	}

	// Strip timestamps from after the end time.
	if pos > 0 && c.res.Timestamps[pos-1] > c.end {
		pos -= 2
		for pos >= 0 && c.res.Timestamps[pos] > c.end {
			pos--
		}
		pos++
	}

	c.res.Timestamps = c.res.Timestamps[:pos]
	c.res.Values = c.res.Values[:pos]

	return c.res
}

func (c *booleanArrayAscendingCursor) nextTSM() *tsdb.BooleanArray {
	c.tsm.keyCursor.Next()
	c.tsm.values, _ = c.tsm.keyCursor.ReadBooleanArrayBlock(c.tsm.buf)
	c.tsm.pos = 0
	return c.tsm.values
}

type booleanArrayDescendingCursor struct {
	cache struct {
		values Values
		pos    int
	}

	tsm struct {
		buf       *tsdb.BooleanArray
		values    *tsdb.BooleanArray
		pos       int
		keyCursor *KeyCursor
	}

	end int64
	res *tsdb.BooleanArray
}

func newBooleanArrayDescendingCursor() *booleanArrayDescendingCursor {
	c := &booleanArrayDescendingCursor{
		res: tsdb.NewBooleanArrayLen(tsdb.DefaultMaxPointsPerBlock),
	}
	c.tsm.buf = tsdb.NewBooleanArrayLen(tsdb.DefaultMaxPointsPerBlock)
	return c
}

func (c *booleanArrayDescendingCursor) reset(seek, end int64, cacheValues Values, tsmKeyCursor *KeyCursor) {
	// Search for the time value greater than the seek time (not included)
	// and then move our position back one which will include the values in
	// our time range.
	c.end = end
	c.cache.values = cacheValues
	c.cache.pos = sort.Search(len(c.cache.values), func(i int) bool {
		return c.cache.values[i].UnixNano() > seek
	})
	c.cache.pos--

	c.tsm.keyCursor = tsmKeyCursor
	c.tsm.values, _ = c.tsm.keyCursor.ReadBooleanArrayBlock(c.tsm.buf)
	c.tsm.pos = sort.Search(c.tsm.values.Len(), func(i int) bool {
		return c.tsm.values.Timestamps[i] > seek
	})
	c.tsm.pos--
}

func (c *booleanArrayDescendingCursor) Err() error { return nil }

func (c *booleanArrayDescendingCursor) Stats() tsdb.CursorStats {
	return tsdb.CursorStats{}
}

func (c *booleanArrayDescendingCursor) Close() {
	if c.tsm.keyCursor != nil {
		c.tsm.keyCursor.Close()
		c.tsm.keyCursor = nil
	}
	c.cache.values = nil
	c.tsm.values = nil
}

func (c *booleanArrayDescendingCursor) Next() *tsdb.BooleanArray {
	pos := 0
	cvals := c.cache.values
	tvals := c.tsm.values

	c.res.Timestamps = c.res.Timestamps[:cap(c.res.Timestamps)]
	c.res.Values = c.res.Values[:cap(c.res.Values)]

	for pos < len(c.res.Timestamps) && c.tsm.pos >= 0 && c.cache.pos >= 0 {
		ckey := cvals[c.cache.pos].UnixNano()
		tkey := tvals.Timestamps[c.tsm.pos]
		if ckey == tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(BooleanValue).value
			c.cache.pos--
			c.tsm.pos--
		} else if ckey > tkey {
			c.res.Timestamps[pos] = ckey
			c.res.Values[pos] = cvals[c.cache.pos].(BooleanValue).value
			c.cache.pos--
		} else {
			c.res.Timestamps[pos] = tkey
			c.res.Values[pos] = tvals.Values[c.tsm.pos]
			c.tsm.pos--
		}

		pos++

		if c.tsm.pos < 0 {
			tvals = c.nextTSM()
		}
	}

	if pos < len(c.res.Timestamps) {
		// cache was exhausted
		if c.tsm.pos >= 0 {
			for pos < len(c.res.Timestamps) && c.tsm.pos >= 0 {
				c.res.Timestamps[pos] = tvals.Timestamps[c.tsm.pos]
				c.res.Values[pos] = tvals.Values[c.tsm.pos]
				pos++
				c.tsm.pos--
				if c.tsm.pos < 0 {
					tvals = c.nextTSM()
				}
			}
		}

		if c.cache.pos >= 0 {
			// TSM was exhausted
			for pos < len(c.res.Timestamps) && c.cache.pos >= 0 {
				c.res.Timestamps[pos] = cvals[c.cache.pos].UnixNano()
				c.res.Values[pos] = cvals[c.cache.pos].(BooleanValue).value
				pos++
				c.cache.pos--
			}
		}
	}

	// Strip timestamps from before the end time.
	if pos > 0 && c.res.Timestamps[pos-1] < c.end {
		pos -= 2
		for pos >= 0 && c.res.Timestamps[pos] < c.end {
			pos--
		}
		pos++
	}

	c.res.Timestamps = c.res.Timestamps[:pos]
	c.res.Values = c.res.Values[:pos]

	return c.res
}

func (c *booleanArrayDescendingCursor) nextTSM() *tsdb.BooleanArray {
	c.tsm.keyCursor.Next()
	c.tsm.values, _ = c.tsm.keyCursor.ReadBooleanArrayBlock(c.tsm.buf)
	c.tsm.pos = len(c.tsm.values.Timestamps) - 1
	return c.tsm.values
}
