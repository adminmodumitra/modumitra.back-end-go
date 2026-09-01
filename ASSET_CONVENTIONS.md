# Living-room curated asset conventions

> **Canonical import requirements:** Product Book
> [Section 36A — 3D Asset Import and Personal Catalog](./CABINET_STUDIO_PRODUCT_BOOK.md#36a-3d-asset-import-and-personal-catalog).
> Cloud storage, processing, and organization ownership are defined in the
> [Backend, SaaS and Commercial Platform Book](./BACKEND_SAAS_COMMERCIAL_PLATFORM_BOOK.md#9-cloud-3d-asset-and-texture-service).

This pack powers soft-goods presentation in Render Studio / Model View.
Cabinets and millwork stay procedural for shop accuracy.

## Layout

```text
public/
  models/soft-goods/
    sofa-3-seat.glb
    lounge-chair.glb
    coffee-table.glb
    side-table.glb
    floor-lamp.glb
    indoor-plant.glb
  textures/
    wood/   oak-*.png  walnut-*.png
    fabric/ oatmeal-color.png olive-color.png rug-wool-color.png
    paint/  wall-color.png
    metal/  charcoal-ao.png
```

Regenerate with:

```bash
node scripts/curated-assets/generate-pack.mjs
```

## GLB rules

| Rule | Convention |
|---|---|
| Units | Metres inside GLB |
| Size | Matches catalog `nativeSizeMm` after normalize |
| Origin | Floor contact at `Y = 0`, centered in XZ |
| Up axis | +Y |
| Forward | +Z toward room interior |
| Mesh names | `{slot}_{part}` e.g. `upholstery_seat`, `legs_-0.88_-0.3` |
| Slots | Must match `modelManifest.materialGroups` tokens |
| Fallback | Procedural adapters always remain available |

### Slot map

| Asset | Slots |
|---|---|
| sofa-3-seat | `upholstery`, `legs` |
| lounge-chair | `upholstery`, `frame` |
| coffee-table | `top`, `frame` |
| side-table | `top`, `frame` |
| floor-lamp | `frame`, `shade` |
| indoor-plant | `foliage`, `planter` |

## Texture rules

- Paths are registry `assetKey` values only — never stored in InteriorProject JSON.
- Prefer PNG/JPG under `public/textures/...`.
- If a texture `available` flag is false or load fails, procedural canvas maps are used.
- Material entities keep color/roughness/metalness as authoring truth; maps add surface detail.

## Product boundary

- Soft goods / decor: curated GLB-first when `available: true`.
- TV unit, bookcase, rug, mirror, cabinets: procedural / millwork path.
