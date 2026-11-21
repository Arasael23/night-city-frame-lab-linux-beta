#!/usr/bin/env python3
"""Write a Go-linkable COFF object (.syso, amd64) holding Windows resources: icon (+group), manifest, version info.
Usage: mksyso.py out.syso icon.ico manifest.xml FILEVER(a,b,c,d) "key=value;key=value..." """
import struct, sys

def u16z(s):
    return s.encode('utf-16-le') + b'\0\0'

def pad4(b):
    return b + b'\0' * ((4 - len(b) % 4) % 4)

def vs_block(key, value=b'', children=b'', vtype=1, wlen=None):
    """one VERSIONINFO node: wLength, wValueLength, wType, szKey, padding, Value, padding, Children"""
    head = struct.pack('<HHH', 0, 0, vtype) + u16z(key)
    head = pad4(head)
    body = head + value
    body = pad4(body) + children
    wval = wlen if wlen is not None else (len(value) // 2 if vtype == 1 else len(value))
    return struct.pack('<HHH', len(body), wval, vtype) + body[6:]

def version_info(fv, strings):
    a, b, c, d = fv
    ffi = struct.pack('<13I', 0xFEEF04BD, 0x00010000, (a << 16) | b, (c << 16) | d, (a << 16) | b, (c << 16) | d,
                      0x3F, 0, 0x40004, 1, 0, 0, 0)  # VOS_NT_WINDOWS32, VFT_APP
    strs = b''.join(pad4(vs_block(k, u16z(v), vtype=1)) for k, v in strings)
    table = vs_block('040904b0', children=strs)
    sfi = vs_block('StringFileInfo', children=pad4(table))
    var = vs_block('Translation', struct.pack('<HH', 0x409, 1200), vtype=0)
    vfi = vs_block('VarFileInfo', children=pad4(var))
    return vs_block('VS_VERSION_INFO', ffi, pad4(sfi) + pad4(vfi), vtype=0, wlen=len(ffi))

def parse_ico(data):
    _, typ, n = struct.unpack_from('<HHH', data, 0)
    imgs = []
    for i in range(n):
        w, h, cc, r, planes, bpp, size, off = struct.unpack_from('<BBBBHHII', data, 6 + 16 * i)
        imgs.append(((w, h, cc, r, planes, bpp, size), data[off:off + size]))
    return imgs

def build(resources):
    """resources: {type: {id: bytes}} (all ids numeric, language 0x409) -> (.rsrc bytes, reloc offsets)"""
    types = sorted(resources)
    # layout: root dir, type dirs, name dirs, data entries, data
    def dirhdr(n):
        return struct.pack('<IIHHHH', 0, 0, 0, 0, 0, n)
    root_size = 16 + 8 * len(types)
    type_dirs = []
    off = root_size
    for t in types:
        type_dirs.append(off); off += 16 + 8 * len(resources[t])
    name_dirs = {}
    for t in types:
        for i in sorted(resources[t]):
            name_dirs[(t, i)] = off; off += 16 + 8
    entries = {}
    for t in types:
        for i in sorted(resources[t]):
            entries[(t, i)] = off; off += 16
    datas = {}
    for t in types:
        for i in sorted(resources[t]):
            off = (off + 7) & ~7
            datas[(t, i)] = off; off += len(resources[t][i])
    buf = bytearray(off)
    buf[0:16] = dirhdr(len(types))
    for k, t in enumerate(types):
        struct.pack_into('<II', buf, 16 + 8 * k, t, 0x80000000 | type_dirs[k])
        buf[type_dirs[k]:type_dirs[k] + 16] = dirhdr(len(resources[t]))
        for j, i in enumerate(sorted(resources[t])):
            struct.pack_into('<II', buf, type_dirs[k] + 16 + 8 * j, i, 0x80000000 | name_dirs[(t, i)])
            nd = name_dirs[(t, i)]
            buf[nd:nd + 16] = dirhdr(1)
            struct.pack_into('<II', buf, nd + 16, 0x409, entries[(t, i)])
    relocs = []
    for (t, i), e in entries.items():
        d = resources[t][i]
        struct.pack_into('<IIII', buf, e, datas[(t, i)], len(d), 0, 0)
        relocs.append(e)
        buf[datas[(t, i)]:datas[(t, i)] + len(d)] = d
    return bytes(buf), sorted(relocs)

def coff(rsrc, relocs):
    nsec = 1
    hdr_size = 20 + 40 * nsec
    raw_off = hdr_size
    rel_off = raw_off + len(rsrc)
    sym_off = rel_off + 10 * len(relocs)
    out = struct.pack('<HHIIIHH', 0x8664, nsec, 0, sym_off, 1, 0, 0)
    out += struct.pack('<8sIIIIIIHHI', b'.rsrc\0\0\0', 0, 0, len(rsrc), raw_off, rel_off, 0, len(relocs), 0, 0x40000040)
    out += rsrc
    for r in relocs:
        out += struct.pack('<IIH', r, 0, 3)          # IMAGE_REL_AMD64_ADDR32NB against symbol 0 (.rsrc)
    out += struct.pack('<8sIhHBB', b'.rsrc\0\0\0', 0, 1, 0, 3, 0)   # section symbol, IMAGE_SYM_CLASS_STATIC
    out += struct.pack('<I', 4)                       # empty string table
    return out

if __name__ == '__main__':
    out, ico, man, fv, strs = sys.argv[1:6]
    fv = tuple(int(x) for x in fv.split(','))
    strings = [tuple(kv.split('=', 1)) for kv in strs.split(';') if kv]
    res = {3: {}, 14: {}, 24: {}, 16: {}}
    grp = struct.pack('<HHH', 0, 1, 0)
    imgs = parse_ico(open(ico, 'rb').read())
    for n, ((w, h, cc, r, planes, bpp, size), data) in enumerate(imgs, 1):
        res[3][n] = data
        grp += struct.pack('<BBBBHHIH', w, h, cc, r, planes, bpp, size, n)
    grp = grp[:4] + struct.pack('<H', len(imgs)) + grp[6:]
    res[14][1] = grp
    res[24][1] = open(man, 'rb').read()
    res[16][1] = version_info(fv, strings)
    rsrc, rel = build(res)
    open(out, 'wb').write(coff(rsrc, rel))
    print('wrote', out, len(rsrc), 'bytes of resources,', len(imgs), 'icon images')
