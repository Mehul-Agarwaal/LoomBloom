CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    owner_name TEXT NOT NULL DEFAULT '',
    phone TEXT NOT NULL DEFAULT '',
    subscription_plan TEXT NOT NULL DEFAULT 'starter',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    plan TEXT NOT NULL DEFAULT 'starter',
    status TEXT NOT NULL DEFAULT 'trialing',
    starts_on DATE NOT NULL DEFAULT CURRENT_DATE,
    ends_on DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE machines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    machine_number TEXT NOT NULL,
    size TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (organization_id, machine_number)
);

CREATE TABLE workers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    phone TEXT NOT NULL DEFAULT '',
    photo_url TEXT NOT NULL DEFAULT '',
    average_production NUMERIC(12, 2) NOT NULL DEFAULT 0,
    close_per_shift INTEGER NOT NULL DEFAULT 0,
    number_of_shifts INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE stock_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    quality TEXT NOT NULL DEFAULT '',
    stored_on TEXT NOT NULL DEFAULT '',
    quantity NUMERIC(12, 2) NOT NULL DEFAULT 0,
    unit TEXT NOT NULL DEFAULT 'meters',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE spare_parts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    quantity NUMERIC(12, 2) NOT NULL DEFAULT 0,
    unit TEXT NOT NULL DEFAULT 'pieces',
    reorder_level NUMERIC(12, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE raw_materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    quality TEXT NOT NULL DEFAULT '',
    quantity NUMERIC(12, 2) NOT NULL DEFAULT 0,
    unit TEXT NOT NULL DEFAULT 'kg',
    stored_on TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    worker_id UUID REFERENCES workers(id) ON DELETE SET NULL,
    amount NUMERIC(12, 2) NOT NULL,
    payment_date DATE NOT NULL DEFAULT CURRENT_DATE,
    mode TEXT NOT NULL DEFAULT 'cash',
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE production_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    machine_id UUID NOT NULL REFERENCES machines(id) ON DELETE CASCADE,
    worker_id UUID REFERENCES workers(id) ON DELETE SET NULL,
    entry_date DATE NOT NULL DEFAULT CURRENT_DATE,
    shift TEXT NOT NULL DEFAULT '',
    quantity NUMERIC(12, 2) NOT NULL DEFAULT 0,
    unit TEXT NOT NULL DEFAULT 'meters',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_machines_organization_id ON machines(organization_id);
CREATE INDEX idx_workers_organization_id ON workers(organization_id);
CREATE INDEX idx_stock_items_organization_id ON stock_items(organization_id);
CREATE INDEX idx_spare_parts_organization_id ON spare_parts(organization_id);
CREATE INDEX idx_raw_materials_organization_id ON raw_materials(organization_id);
CREATE INDEX idx_payments_organization_id ON payments(organization_id);
CREATE INDEX idx_production_entries_org_date ON production_entries(organization_id, entry_date DESC);
