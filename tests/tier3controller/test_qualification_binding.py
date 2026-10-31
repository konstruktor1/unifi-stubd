"""Test role/SSH validation with contained synthetic inventory, without a VM."""

import json
import shutil
import tempfile
import unittest
from pathlib import Path

from lab.stub.tools.qualification_binding import validate_binding


class QualificationBindingTests(unittest.TestCase):
    """Transport inputs must bind the assigned product source and trust file."""

    @classmethod
    def setUpClass(cls):
        cls.root = Path(__file__).resolve().parents[2]
        cls.work = cls.root / "tests/.work"
        cls.created_work = not cls.work.exists()
        cls.work.mkdir(mode=0o700, exist_ok=True)
        if cls.work.resolve(strict=True) != cls.work:
            raise ValueError("test fixture directory must be physical")

    @classmethod
    def tearDownClass(cls):
        if cls.created_work:
            cls.work.rmdir()

    def setUp(self):
        self.fixture = Path(tempfile.mkdtemp(prefix="product-binding-", dir=self.work))
        self.addCleanup(shutil.rmtree, self.fixture)
        self.data = self.fixture / "data"
        self.machine = self.data / "tier3/machines/unifi-stubd-lab"
        self.known_hosts = self.machine / "identity/known_hosts.test-runner"
        self.identity = self.fixture / "fixture-identity"
        self.seed = self.machine / "base/seed.iso"
        for file in (self.known_hosts, self.identity, self.seed):
            file.parent.mkdir(parents=True, exist_ok=True)
            file.write_text(
                "synthetic fixture, no key or machine data\n", encoding="utf-8"
            )
        self.inventory = (
            self.fixture
            / ".consumer-data/build-runner/test-runner/inventory/product-source.toml"
        )
        self.inventory.parent.mkdir(parents=True)
        self.source = (
            self.root / "tests/fixtures/tier3controller/product-inventory.toml"
        ).read_text(encoding="utf-8")
        for key, value in {
            "SOURCE": str(self.data / "tier3/test-runner/_managed/product-source"),
            "IDENTITY": str(self.identity),
            "KNOWN_HOSTS": str(self.known_hosts),
            "DRIVE": f"if=none,format=raw,readonly=on,file={self.seed}",
        }.items():
            self.source = self.source.replace(f"@{key}@", json.dumps(value))
        self.inventory.write_text(self.source, encoding="utf-8")
        self.transport = {
            "host": "127.0.0.1",
            "port": 2222,
            "user": "fixture",
            "identity_file": str(self.identity),
            "known_hosts_file": str(self.known_hosts),
        }

    def check_binding(self):
        validate_binding(self.fixture, self.data, "product-source", self.transport)

    def test_native_product_binding(self):
        self.check_binding()

    def test_ssh_port_must_match_native_transport(self):
        self.transport["port"] = 2223
        with self.assertRaisesRegex(ValueError, "native product transport"):
            self.check_binding()

    def test_controller_source_cannot_replace_product(self):
        self.inventory.write_text(
            self.source.replace(
                "/_managed/product-source", "/_managed/controller-source"
            ),
            encoding="utf-8",
        )
        with self.assertRaisesRegex(ValueError, "assigned ARM64 product role"):
            self.check_binding()

    def test_host_key_checking_cannot_be_disabled(self):
        self.inventory.write_text(
            self.source.replace(
                'host_key_checking = "yes"', 'host_key_checking = "no"'
            ),
            encoding="utf-8",
        )
        with self.assertRaisesRegex(ValueError, "native product transport"):
            self.check_binding()

    def test_other_machine_seed_cannot_replace_product(self):
        self.inventory.write_text(
            self.source.replace("/base/seed.iso", "/base/other-machine.iso"),
            encoding="utf-8",
        )
        with self.assertRaisesRegex(
            ValueError, "assigned product machine seed binding"
        ):
            self.check_binding()

    def test_inventory_symlink_rejected_without_changing_target(self):
        target = self.fixture / "sentinel.toml"
        target.write_text(self.source, encoding="utf-8")
        self.inventory.unlink()
        self.inventory.symlink_to(target)
        with self.assertRaisesRegex(ValueError, "bounded physical file"):
            self.check_binding()
        self.assertEqual(target.read_text(encoding="utf-8"), self.source)

    def test_malformed_native_schema_rejected(self):
        self.inventory.write_text('hosts = ["unexpected"]\n', encoding="utf-8")
        with self.assertRaisesRegex(ValueError, "assigned ARM64 product role"):
            self.check_binding()
